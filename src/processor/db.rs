use dashmap::mapref::one::{Ref, RefMut};
use dashmap::DashMap;
use std::collections::HashSet;
use std::io;

use crate::command::Command;
use crate::processor::spec::Processor;
use crate::response::Response;
use crate::wal::WAL;

pub(crate) enum DbValue {
    String(String),
    List(Vec<String>),
    Set(HashSet<String>),
}

pub(crate) struct InMemoryDb {
    pub(crate) data: DashMap<String, DbValue>,
}

impl InMemoryDb {
    pub(crate) fn new() -> io::Result<InMemoryDb> {
        Ok(InMemoryDb {
            data: DashMap::new(),
        })
    }

    fn exists(&self, key: &String) -> Response {
        let found = if self.data.contains_key(key) { 1 } else { 0 };
        Response::Integer(found)
    }

    pub(crate) fn delete(&self, keys: &Vec<String>) -> Response {
        let deleted_rows = keys.into_iter().fold(0, |acc, value| {
            acc + match self.data.remove(&value.to_string()) {
                Some(_) => 1,
                None => 0,
            }
        });
        Response::Integer(deleted_rows)
    }

    pub(crate) fn get_value_ref<F>(
        &self,
        key: &String,
        is_variant: F,
    ) -> Option<Ref<String, DbValue>>
    where
        F: Fn(&DbValue) -> bool,
    {
        match self.data.get(key) {
            Some(item) if is_variant(item.value()) => Some(item),
            _ => None,
        }
    }

    pub(crate) fn get_value_ref_mut<F>(
        &self,
        key: &String,
        is_variant: F,
    ) -> Option<RefMut<String, DbValue>>
    where
        F: Fn(&DbValue) -> bool,
    {
        match self.data.get_mut(key) {
            Some(item) if is_variant(item.value()) => Some(item),
            _ => None,
        }
    }
}

macro_rules! log_if_some {
    ($var:expr, $cmd:expr) => {
        if let Some(var) = $var {
            var.log(&$cmd).unwrap();
        }
    };
}

impl Processor for InMemoryDb {
    fn cmd(&self, cmd: Command, wal: Option<&mut WAL>) -> Response {
        match cmd {
            Command::Get { ref key } => self.get(key),
            Command::GetDel { ref key } => {
                log_if_some!(wal, cmd);
                self.get_del(key)
            }
            Command::Set { ref key, ref value } => {
                log_if_some!(wal, cmd);
                self.set(key, value)
            }
            Command::Delete { ref keys } => {
                log_if_some!(wal, cmd);
                self.delete(keys)
            }
            Command::LPush {
                ref key,
                ref values,
            } => {
                log_if_some!(wal, cmd);
                self.lpush(key, values)
            }
            Command::RPush {
                ref key,
                ref values,
            } => {
                log_if_some!(wal, cmd);
                self.rpush(key, values)
            }
            Command::LPop { ref key } => {
                log_if_some!(wal, cmd);
                self.lpop(key)
            }
            Command::RPop { ref key } => {
                log_if_some!(wal, cmd);
                self.rpop(key)
            }
            Command::LRange {
                ref key,
                start,
                end,
            } => self.lrange(key, start, end),
            Command::LLen { ref key } => self.llen(key),
            Command::Exists { ref key } => self.exists(key),
            Command::Incr { ref key } => {
                log_if_some!(wal, cmd);
                self.incr(key)
            }
            Command::IncrBy {
                ref key,
                ref increment,
            } => {
                log_if_some!(wal, cmd);
                self.incr_by(key, increment)
            }
            Command::Decr { ref key } => {
                log_if_some!(wal, cmd);
                self.decr(key)
            }
            Command::DecrBy {
                ref key,
                ref decrement,
            } => {
                log_if_some!(wal, cmd);
                self.decr_by(key, decrement)
            }
            Command::SAdd {
                ref key,
                ref values,
            } => {
                log_if_some!(wal, cmd);
                self.sadd(&key, &values)
            }
            Command::SCard { ref key } => self.scard(&key),
            Command::SInter {
                ref key,
                ref others,
            } => self.sinter(key, others),
            Command::SDiff {
                ref key,
                ref others,
            } => self.sdiff(key, others),
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use crate::response::Error::UnknownKey;
    use crate::response::OK;

    #[test]
    fn test_insert() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");
        let response = db.set(&key, &value);

        assert_eq!(response, Response::String(String::from(OK)));
    }

    #[test]
    fn test_get() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let set_response = db.set(&key, &value);
        assert_eq!(set_response, Response::String(String::from(OK)));

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::String(value));
    }

    #[test]
    fn test_delete() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let set_response = db.set(&key, &value);
        assert_eq!(set_response, Response::String(String::from(OK)));

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::String(value));

        let delete_response = db.delete(&[String::from("name")].to_vec());
        assert_eq!(delete_response, Response::Integer(1));

        let get_response = db.get(&String::from("name"));
        assert_eq!(get_response, Response::Err(UnknownKey));
    }

    #[test]
    fn test_list() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("fruits");

        let apple = String::from("apple");
        let mango = String::from("mango");
        let orange = String::from("orange");

        let apple_lpush_response = db.lpush(&key, &[apple].to_owned().to_vec());
        assert_eq!(apple_lpush_response, Response::Integer(1));

        let llen_response = db.llen(&key);
        assert_eq!(llen_response, Response::Integer(1))
    }
}