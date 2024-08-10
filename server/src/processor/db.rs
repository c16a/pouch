use crate::processor::spec::Processor;
use crate::structures::sorted_set::SortedSet;
use crate::wal::WAL;
use dashmap::mapref::one::{Ref, RefMut};
use dashmap::DashMap;
use pouch_sdk::command::Command;
use pouch_sdk::response::Response;
use std::collections::HashSet;
use std::io;

pub(crate) enum DbValue {
    String {
        value: String,
        expiry_ts: u64,
    },
    List(Vec<String>),
    Set(HashSet<String>),
    SortedSet(SortedSet<String>),
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
        Response::BooleanValue {
            value: self.data.contains_key(key),
        }
    }

    pub(crate) fn delete(&self, keys: &Vec<String>) -> Response {
        let deleted_rows = keys.into_iter().fold(0, |acc, value| {
            acc + match self.data.remove(&value.to_string()) {
                Some(_) => 1,
                None => 0,
            }
        });
        Response::AffectedKeys {
            affected_keys: deleted_rows,
        }
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
            Command::Set { ref key, ref value, ref expiry_seconds } => {
                log_if_some!(wal, cmd);
                self.set(key, value, expiry_seconds)
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
            Command::ZAdd {
                ref key,
                ref values,
            } => {
                log_if_some!(wal, cmd);
                self.zadd(&key, values)
            }
            Command::ZCard { ref key } => self.zcard(&key),
        }
    }
}

#[cfg(test)]
mod test {
    use super::*;
    use pouch_sdk::response::Error::UnknownKey;
    use std::time::{SystemTime, UNIX_EPOCH};

    #[test]
    fn test_insert() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");
        let expiry_seconds = SystemTime::now().duration_since(UNIX_EPOCH).expect("").as_secs() + 100;
        let response = db.set(&key, &value, &expiry_seconds);

        assert_eq!(response, Response::AffectedKeys { affected_keys: 1 });
    }

    #[test]
    fn test_get() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let expiry_seconds = SystemTime::now().duration_since(UNIX_EPOCH).expect("").as_secs() + 100;
        let set_response = db.set(&key, &value, &expiry_seconds);
        assert_eq!(set_response, Response::AffectedKeys { affected_keys: 1 });

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::StringValue { value });
    }

    #[test]
    fn test_delete() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let expiry_seconds = SystemTime::now().duration_since(UNIX_EPOCH).expect("").as_secs() + 100;
        let set_response = db.set(&key, &value, &expiry_seconds);
        assert_eq!(set_response, Response::AffectedKeys { affected_keys: 1 });

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::StringValue { value });

        let delete_response = db.delete(&[String::from("name")].to_vec());
        assert_eq!(delete_response, Response::AffectedKeys { affected_keys: 1 });

        let get_response = db.get(&String::from("name"));
        assert_eq!(get_response, Response::Err { error: UnknownKey });
    }

    #[test]
    fn test_list() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("fruits");

        let apple = String::from("apple");
        let mango = String::from("mango");
        let orange = String::from("orange");

        let apple_lpush_response = db.lpush(&key, &[apple].to_owned().to_vec());
        assert_eq!(
            apple_lpush_response,
            Response::AffectedKeys { affected_keys: 1 }
        );

        let llen_response = db.llen(&key);
        assert_eq!(llen_response, Response::Count { count: 1 });
    }
}
