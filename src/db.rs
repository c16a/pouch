use std::collections::hash_map::Entry;
use std::collections::HashMap;
use std::io;
use crate::response::Response;

#[derive(Debug, Clone)]
enum DbValue {
    String(String),
    List(Vec<String>),
}

#[derive(Clone)]
pub(crate) struct InMemoryDb {
    data: HashMap<String, DbValue>,
}

impl InMemoryDb {
    pub(crate) fn new() -> io::Result<InMemoryDb> {
        Ok(InMemoryDb { data: HashMap::new() })
    }

    pub(crate) fn get(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value {
                DbValue::String(value) => {
                    Response::SimpleString { value: value.to_string() }
                }
                DbValue::List(_) => {
                    Response::SimpleString { value: String::from("(nil)") }
                }
            }
        } else {
            Response::SimpleString { value: String::from("(nil)") }
        }
    }

    pub(crate) fn insert(&mut self, key: &String, value: &String) -> Response {
        self.data.insert(key.to_string(), DbValue::String(value.to_string()));
        Response::SimpleString { value: String::from("OK") }
    }

    pub(crate) fn remove(&mut self, key: &String) -> Response {
        self.data.remove(&key.to_string());
        Response::SimpleString { value: String::from("OK") }
    }

    pub(crate) fn lpush(&mut self, key: &String, value: &String) -> Response {
        match self.data.entry(key.to_string()) {
            Entry::Occupied(mut entry) => {
                if let DbValue::List(list) = entry.get_mut() {
                    list.insert(0, value.to_string());
                }
            }
            Entry::Vacant(entry) => {
                entry.insert(DbValue::List(vec![value.to_string()]));
            }
        }
        Response::SimpleString { value: String::from("OK") }
    }

    pub(crate) fn rpush(&mut self, key: &String, value: &String) -> Response {
        match self.data.entry(key.to_string()) {
            Entry::Occupied(mut entry) => {
                if let DbValue::List(list) = entry.get_mut() {
                    list.push(value.to_string());
                }
            }
            Entry::Vacant(entry) => {
                entry.insert(DbValue::List(vec![value.to_string()]));
            }
        }
        Response::SimpleString { value: String::from("OK") }
    }

    pub(crate) fn lrange(&self, key: &String, start: Option<usize>, end: Option<usize>) -> Response {
        match self.data.get(&key.clone()) {
            Some(DbValue::List(list)) => {
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
            _ => Response::SimpleString { value: "(nil)".to_string() },
        }
    }
}