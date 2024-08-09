use crate::processor::db::{DbValue, InMemoryDb};
use dashmap::mapref::one::{Ref, RefMut};
use pouch_sdk::response::Error::{IncompatibleDataType, UnknownKey};
use pouch_sdk::response::Response;

impl InMemoryDb {
    fn get_list_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::List(_)))
    }
    fn get_list_ref_mut(&self, key: &String) -> Option<RefMut<String, DbValue>> {
        self.get_value_ref_mut(key, |value| matches!(value, DbValue::List(_)))
    }
    pub(crate) fn lpush(&self, key: &String, values: &Vec<String>) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    values.into_iter().for_each(|value| {
                        list.insert(0, value.to_string());
                    });
                    Response::Integer(list.len() as i64)
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => {
                let mut list = Vec::new();
                values.into_iter().for_each(|value| {
                    list.insert(0, value.to_string());
                });
                self.data.insert(key.to_string(), DbValue::List(list));
                Response::Integer(values.len() as i64)
            }
        }
    }

    pub(crate) fn rpush(&self, key: &String, values: &Vec<String>) -> Response {
        match self.data.get_mut(&key.clone()) {
            Some(mut db_value) => match db_value.value_mut() {
                DbValue::List(list) => {
                    values
                        .into_iter()
                        .for_each(|value| list.push(value.to_string()));
                    Response::Integer(list.len() as i64)
                }
                _ => Response::Err(IncompatibleDataType),
            },
            None => {
                let list = values.to_vec();
                self.data.insert(key.to_string(), DbValue::List(list));
                Response::Integer(values.len() as i64)
            }
        }
    }

    pub(crate) fn lpop(&self, key: &String) -> Response {
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

    pub(crate) fn rpop(&self, key: &String) -> Response {
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

    pub(crate) fn lrange(
        &self,
        key: &String,
        start: Option<usize>,
        end: Option<usize>,
    ) -> Response {
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

    pub(crate) fn llen(&self, key: &String) -> Response {
        if let Some(list_ref) = self.get_list_ref(&key) {
            if let DbValue::List(list) = list_ref.value() {
                let len = list.len();
                Response::Integer(len as i64)
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }
}
