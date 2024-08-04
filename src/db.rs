use std::collections::HashSet;
use std::io;

use dashmap::DashMap;

use crate::command::Command;
use crate::processor::Processor;
use crate::response::{Response, FALSE, INCOMPATIBLE_DATA_TYPE, OK, TRUE, UNKNOWN_KEY};
use crate::wal::WAL;

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
                DbValue::String(value) => Response::SimpleString {
                    value: value.to_string(),
                },
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            }
        } else {
            Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            }
        }
    }

    fn exists(&self, key: &String) -> Response {
        let found = if self.data.contains_key(key) {
            TRUE
        } else {
            FALSE
        };
        Response::SimpleString {
            value: String::from(found),
        }
    }

    fn set(&self, key: &String, value: &String) -> Response {
        self.data
            .insert(key.to_string(), DbValue::String(value.to_string()));
        Response::SimpleString {
            value: String::from(OK),
        }
    }

    fn remove(&self, key: &String) -> Response {
        self.data.remove(&key.to_string());
        Response::SimpleString {
            value: String::from(OK),
        }
    }

    fn lpush(&self, key: &String, value: &String) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    list.insert(0, value.to_string());
                    Response::SimpleString {
                        value: String::from(OK),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => {
                self.data
                    .insert(key.to_string(), DbValue::List(vec![value.to_string()]));
                Response::SimpleString {
                    value: String::from(OK),
                }
            }
        }
    }

    fn rpush(&self, key: &String, value: &String) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    list.push(value.to_string());
                    Response::SimpleString {
                        value: String::from(OK),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => {
                self.data
                    .insert(key.to_string(), DbValue::List(vec![value.to_string()]));
                Response::SimpleString {
                    value: String::from(OK),
                }
            }
        }
    }

    fn lpop(&self, key: &String) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    let el = list.remove(0);
                    Response::SimpleString {
                        value: String::from(el),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            },
        }
    }

    fn rpop(&self, key: &String) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    let len = list.len();
                    let el = list.remove(len - 1);
                    Response::SimpleString {
                        value: String::from(el),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            },
        }
    }

    fn lrange(&self, key: &String, start: Option<usize>, end: Option<usize>) -> Response {
        match self.data.get(&key.clone()) {
            Some(db_value) => match db_value.value() {
                DbValue::List(list) => {
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
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            },
        }
    }

    fn llen(&self, key: &String) -> Response {
        match self.data.get(&key.clone()) {
            Some(db_value) => match db_value.value() {
                DbValue::List(list) => {
                    let len = list.len();
                    Response::SimpleString {
                        value: len.to_string(),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            },
        }
    }

    fn sadd(&self, key: &String, values: &Vec<String>) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::Set(set) => {
                    values.into_iter().for_each(|value| {
                        set.insert(value.to_string());
                    });
                    Response::SimpleString {
                        value: String::from(OK),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => {
                let mut set = HashSet::new();
                values.into_iter().for_each(|value| {
                    set.insert(value.to_string());
                });
                self.data
                    .insert(key.to_string(), DbValue::Set(set));
                Response::SimpleString {
                    value: String::from(OK),
                }
            }
        }
    }

    fn scard(&self, key: &String) -> Response {
        match self.data.get(&key.clone()) {
            Some(db_value) => match db_value.value() {
                DbValue::Set(set) => {
                    let len = set.len();
                    Response::SimpleString {
                        value: len.to_string(),
                    }
                }
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            },
            None => Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            },
        }
    }

    fn incr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i32>() {
                    Ok(x) => {
                        let y = x + 1;
                        self.set(key, &y.to_string());
                        Response::SimpleString {
                            value: String::from(OK),
                        }
                    }
                    Err(_err) => Response::SimpleString {
                        value: String::from(INCOMPATIBLE_DATA_TYPE),
                    },
                },
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            }
        } else {
            Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            }
        }
    }

    fn decr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i32>() {
                    Ok(x) => {
                        let y = x - 1;
                        self.set(key, &y.to_string());
                        Response::SimpleString {
                            value: String::from(OK),
                        }
                    }
                    Err(_err) => Response::SimpleString {
                        value: String::from(INCOMPATIBLE_DATA_TYPE),
                    },
                },
                _ => Response::SimpleString {
                    value: String::from(INCOMPATIBLE_DATA_TYPE),
                },
            }
        } else {
            Response::SimpleString {
                value: String::from(UNKNOWN_KEY),
            }
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
            Command::Delete { ref key } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.remove(key)
            }
            Command::LPush { ref key, ref value } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.lpush(key, value)
            }
            Command::RPush { ref key, ref value } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.rpush(key, value)
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
            Command::SAdd { ref key, ref values } => {
                if let Some(wal) = wal {
                    wal.log(&cmd).unwrap()
                }
                self.sadd(&key, &values)
            }
            Command::SCard { ref key } => {
                self.scard(&key)
            }
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
            Response::SimpleString {
                value: String::from(OK)
            }
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
            Response::SimpleString {
                value: String::from(OK)
            }
        );

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::SimpleString { value });
    }

    #[test]
    fn test_delete() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("name");
        let value = String::from("c16a");

        let set_response = db.set(&key, &value);
        assert_eq!(
            set_response,
            Response::SimpleString {
                value: String::from(OK)
            }
        );

        let get_response = db.get(&key);
        assert_eq!(get_response, Response::SimpleString { value });

        let delete_response = db.remove(&key);
        assert_eq!(
            delete_response,
            Response::SimpleString {
                value: String::from(OK)
            }
        );

        let get_response = db.get(&key);
        assert_eq!(
            get_response,
            Response::SimpleString {
                value: String::from(UNKNOWN_KEY)
            }
        );
    }

    #[test]
    fn test_list() {
        let db = InMemoryDb::new().unwrap();

        let key = String::from("fruits");

        let apple = String::from("apple");
        let mango = String::from("mango");
        let orange = String::from("orange");

        let apple_lpush_response = db.lpush(&key, &apple);
        assert_eq!(
            apple_lpush_response,
            Response::SimpleString {
                value: String::from(OK)
            }
        );
    }
}
