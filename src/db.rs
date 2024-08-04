use dashmap::Entry;
use std::io;
use dashmap::DashMap;
use crate::processor::Processor;
use crate::response::{FALSE, INCOMPATIBLE_DATA_TYPE, OK, Response, TRUE, UNKNOWN_KEY};

#[derive(Debug, Clone)]
enum DbValue {
    String(String),
    List(Vec<String>),
}

#[derive(Clone)]
pub(crate) struct InMemoryDb {
    data: DashMap<String, DbValue>,
}

impl InMemoryDb {
    pub(crate) fn new() -> io::Result<InMemoryDb> {
        Ok(InMemoryDb { data: DashMap::new() })
    }
}

impl Processor for InMemoryDb {
    fn get(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => {
                    Response::SimpleString { value: value.to_string() }
                }
                DbValue::List(_) => {
                    Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                }
            }
        } else {
            Response::SimpleString { value: String::from(UNKNOWN_KEY) }
        }
    }


    fn exists(&self, key: &String) -> Response {
        let found = if self.data.contains_key(key) { TRUE } else { FALSE };
        Response::SimpleString { value: String::from(found) }
    }

    fn set(&self, key: &String, value: &String) -> Response {
        self.data.insert(key.to_string(), DbValue::String(value.to_string()));
        Response::SimpleString { value: String::from(OK) }
    }

    fn remove(&self, key: &String) -> Response {
        self.data.remove(&key.to_string());
        Response::SimpleString { value: String::from(OK) }
    }

    fn lpush(&self, key: &String, value: &String) -> Response {
        match self.data.entry(key.to_string()) {
            Entry::Occupied(mut entry) => {
                match entry.get_mut() {
                    DbValue::List(list) => {
                        list.insert(0, value.to_string());
                        Response::SimpleString { value: String::from(OK) }
                    }
                    _ => {
                        Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                    }
                }
            }
            Entry::Vacant(entry) => {
                entry.insert(DbValue::List(vec![value.to_string()]));
                Response::SimpleString { value: String::from(OK) }
            }
        }
    }

    fn rpush(&self, key: &String, value: &String) -> Response {
        match self.data.entry(key.to_string()) {
            Entry::Occupied(mut entry) => {
                match entry.get_mut() {
                    DbValue::List(list) => {
                        list.push(value.to_string());
                        Response::SimpleString { value: String::from(OK) }
                    }
                    _ => {
                        Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                    }
                }
            }
            Entry::Vacant(entry) => {
                entry.insert(DbValue::List(vec![value.to_string()]));
                Response::SimpleString { value: String::from(OK) }
            }
        }
    }

    fn lrange(&self, key: &String, start: Option<usize>, end: Option<usize>) -> Response {
        match self.data.get(&key.clone()) {
            Some(db_value) => {
                match db_value.value() {
                    DbValue::List(list) => {
                        let len = list.len();
                        let start = match start {
                            Some(val) => val.min(len),
                            None => 0
                        };
                        let end = match end {
                            Some(val) => val.min(len),
                            None => len
                        };
                        let range = &list[start..end];
                        Response::List { values: range.to_vec() }
                    }
                    _ => Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                }
            }
            None => {
                Response::SimpleString { value: String::from(UNKNOWN_KEY) }
            }
        }
    }

    fn incr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => {
                    match value.parse::<i32>() {
                        Ok(x) => {
                            let y = x + 1;
                            self.set(key, &y.to_string());
                            Response::SimpleString { value: String::from(OK) }
                        }
                        Err(_err) => {
                            Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                        }
                    }
                }
                _ => {
                    Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                }
            }
        } else {
            Response::SimpleString { value: String::from(UNKNOWN_KEY) }
        }
    }

    fn decr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => {
                    match value.parse::<i32>() {
                        Ok(x) => {
                            let y = x - 1;
                            self.set(key, &y.to_string());
                            Response::SimpleString { value: String::from(OK) }
                        }
                        Err(_err) => {
                            Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                        }
                    }
                }
                _ => {
                    Response::SimpleString { value: String::from(INCOMPATIBLE_DATA_TYPE) }
                }
            }
        } else {
            Response::SimpleString { value: String::from(UNKNOWN_KEY) }
        }
    }
}