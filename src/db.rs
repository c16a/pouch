use std::collections::HashSet;
use std::io;

use crate::command::Command;
use crate::processor::Processor;
use crate::response::Error::{IncompatibleDataType, NotInteger, UnknownKey};
use crate::response::{Response, OK};
use crate::wal::WAL;
use dashmap::mapref::one::{Ref, RefMut};
use dashmap::DashMap;

enum DbValue {
    String(String),
    List(Vec<String>),
    Set(HashSet<String>),
}

pub(crate) struct InMemoryDb {
    data: DashMap<String, DbValue>,
}

impl InMemoryDb {
    pub(crate) fn new() -> io::Result<InMemoryDb> {
        Ok(InMemoryDb {
            data: DashMap::new(),
        })
    }

    fn get(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => Response::String(value.to_string()),
                _ => Response::Err(IncompatibleDataType),
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn exists(&self, key: &String) -> Response {
        let found = if self.data.contains_key(key) {
            1
        } else {
            0
        };
        Response::Integer(found)
    }

    fn set(&self, key: &String, value: &String) -> Response {
        self.data
            .insert(key.to_string(), DbValue::String(value.to_string()));
        Response::String(String::from(OK))
    }

    fn delete(&self, keys: &Vec<String>) -> Response {
        let deleted_rows = keys.into_iter().fold(0, |acc, value| {
            acc + match self.data.remove(&value.to_string()) {
                Some(_) => 1,
                None => 0
            }
        });
        Response::Integer(deleted_rows)
    }

    fn lpush(&self, key: &String, values: &Vec<String>) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    values.into_iter().for_each(|value| {
                        list.insert(0, value.to_string());
                    });
                    Response::Integer(list.len() as i32)
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => {
                let mut list = Vec::new();
                values.into_iter().for_each(|value| {
                    list.insert(0, value.to_string());
                });
                self.data.insert(key.to_string(), DbValue::List(list));
                Response::Integer(values.len() as i32)
            }
        }
    }

    fn rpush(&self, key: &String, values: &Vec<String>) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    values.into_iter().for_each(|value| {
                        list.push(value.to_string())
                    });
                    Response::Integer(list.len() as i32)
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => {
                let list = values.to_vec();
                self.data.insert(key.to_string(), DbValue::List(list));
                Response::Integer(values.len() as i32)
            }
        }
    }

    fn lpop(&self, key: &String) -> Response {
        if let Some(mut list_ref) = self.get_list_ref_mut(&key) {
            if let DbValue::List(list) = list_ref.value_mut() {
                let el = list.remove(0);
                Response::String(el)
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn rpop(&self, key: &String) -> Response {
        if let Some(mut list_ref) = self.get_list_ref_mut(&key) {
            if let DbValue::List(list) = list_ref.value_mut() {
                let len = list.len();
                let el = list.remove(len - 1);
                Response::String(el)
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn lrange(&self, key: &String, start: Option<usize>, end: Option<usize>) -> Response {
        if let Some(list_ref) = self.get_list_ref(&key) {
            if let DbValue::List(list) = list_ref.value() {
                let len = list.len();
                let start = match start {
                    Some(val) => val.min(len),
                    None => 0,
                };
                let end = match end {
                    Some(val) => val.min(len),
                    None => len,
                };
                let range = &list[start..end];
                Response::List {
                    values: range.to_vec(),
                }
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn llen(&self, key: &String) -> Response {
        if let Some(list_ref) = self.get_list_ref(&key) {
            if let DbValue::List(list) = list_ref.value() {
                let len = list.len();
                Response::Integer(len as i32)
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn sadd(&self, key: &String, values: &Vec<String>) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::Set(set) => {
                    let inserted_rows = values.into_iter().fold(0, |acc, value| {
                        acc + if set.insert(value.to_string()) { 1 } else { 0 }
                    });
                    Response::String(inserted_rows.to_string())
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => {
                let mut set = HashSet::new();
                let inserted_rows = values.into_iter().fold(0, |acc, value| {
                    acc + if set.insert(value.to_string()) { 1 } else { 0 }
                });
                self.data.insert(key.to_string(), DbValue::Set(set));
                Response::String(inserted_rows.to_string())
            }
        }
    }

    fn scard(&self, key: &String) -> Response {
        if let Some(set_ref) = self.get_set_ref(&key) {
            if let DbValue::Set(set) = set_ref.value() {
                let len = set.len();
                Response::String(len.to_string())
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn get_value_ref<F>(&self, key: &String, is_variant: F) -> Option<Ref<String, DbValue>>
    where
        F: Fn(&DbValue) -> bool,
    {
        match self.data.get(key) {
            Some(item) if is_variant(item.value()) => Some(item),
            _ => None,
        }
    }

    fn get_value_ref_mut<F>(&self, key: &String, is_variant: F) -> Option<RefMut<String, DbValue>>
    where
        F: Fn(&DbValue) -> bool,
    {
        match self.data.get_mut(key) {
            Some(item) if is_variant(item.value()) => Some(item),
            _ => None,
        }
    }

    fn get_set_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::Set(_)))
    }

    fn get_list_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::List(_)))
    }

    fn get_list_ref_mut(&self, key: &String) -> Option<RefMut<String, DbValue>> {
        self.get_value_ref_mut(key, |value| matches!(value, DbValue::List(_)))
    }

    fn get_string_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::String(_)))
    }

    fn sinter(&self, key: &String, others: &Vec<String>) -> Response {
        match self.data.get(&key.clone()) {
            Some(db_value) => match db_value.value() {
                DbValue::Set(set) => {
                    let mut intersection = set.clone();
                    for other_key in others {
                        if let Some(other_ref) = self.get_set_ref(other_key) {
                            if let DbValue::Set(other_set) = other_ref.value() {
                                // Perform the intersection and collect into a new HashSet
                                intersection =
                                    intersection.intersection(&other_set).cloned().collect();
                            }
                        } else {
                            // If any other set is missing, the intersection is empty
                            intersection.clear();
                            break;
                        }
                    }
                    Response::Set {
                        values: intersection.into_iter().collect(),
                    }
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => Response::Err(UnknownKey),
        }
    }

    fn sdiff(&self, key: &String, others: &Vec<String>) -> Response {
        match self.data.get(&key.clone()) {
            Some(db_value) => match db_value.value() {
                DbValue::Set(set) => {
                    let mut difference = set.clone();
                    for other_key in others {
                        if let Some(other_ref) = self.get_set_ref(other_key) {
                            if let DbValue::Set(other_set) = other_ref.value() {
                                difference = difference.difference(&other_set).cloned().collect();
                            }
                        }
                    }
                    Response::Set {
                        values: difference.into_iter().collect(),
                    }
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => Response::Err(UnknownKey),
        }
    }

    fn incr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i32>() {
                    Ok(x) => {
                        let y = x + 1;
                        self.set(key, &y.to_string());
                        Response::String(y.to_string())
                    }
                    Err(_err) => Response::Err(NotInteger),
                },
                _ => Response::Err(IncompatibleDataType),
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    fn decr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i32>() {
                    Ok(x) => {
                        let y = x - 1;
                        self.set(key, &y.to_string());
                        Response::String(y.to_string())
                    }
                    Err(_err) => Response::Err(NotInteger),
                },
                _ => Response::Err(IncompatibleDataType),
            }
        } else {
            Response::Err(UnknownKey)
        }
    }
}

impl Processor for InMemoryDb {
    fn cmd(&self, cmd: Command, wal: Option<&mut WAL>) -> Response {
        match cmd {
            Command::Get { ref key } => self.get(key),
            Command::Set { ref key, ref value } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.set(key, value)
            }
            Command::Delete { ref keys } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.delete(keys)
            }
            Command::LPush { ref key, ref values } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.lpush(key, values)
            }
            Command::RPush { ref key, ref values } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.rpush(key, values)
            }
            Command::LPop { ref key } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.lpop(key)
            }
            Command::RPop { ref key } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.rpop(key)
            }
            Command::LRange {
                ref key,
                start,
                end,
            } => self.lrange(key, start, end),
            Command::LLen { ref key } => self.llen(key),
            Command::Exists { ref key } => self.exists(key),
            Command::Incr { ref key } => self.incr(key),
            Command::Decr { ref key } => self.decr(key),
            Command::SAdd {
                ref key,
                ref values,
            } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
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

    #[test]
    fn test_insert() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");
        let response = db.set(&key, &value);

        assert_eq!(
            response,
            Response::String(String::from(OK))
        );
    }

    #[test]
    fn test_get() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let set_response = db.set(&key, &value);
        assert_eq!(
            set_response,
            Response::String(String::from(OK))
        );

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::String(value));
    }

    #[test]
    fn test_delete() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let set_response = db.set(&key, &value);
        assert_eq!(
            set_response,
            Response::String(String::from(OK))
        );

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::String(value));

        let delete_response = db.delete(&[String::from("name")].to_vec());
        assert_eq!(
            delete_response,
            Response::String(String::from(OK))
        );

        let get_response = db.get(&String::from("name"));
        assert_eq!(
            get_response,
            Response::Err(UnknownKey)
        );
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
            Response::Integer(1)
        );

        let llen_response = db.llen(&key);
        assert_eq!(
            llen_response,
            Response::Integer(1)
        )
    }
}
