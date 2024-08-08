use crate::processor::db::{DbValue, InMemoryDb};
use crate::response::Error::{IncompatibleDataType, NotInteger, UnknownKey};
use crate::response::{Response, OK};
use dashmap::mapref::one::Ref;

impl InMemoryDb {
    pub(crate) fn get(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => Response::String(value.to_string()),
                _ => Response::Err(IncompatibleDataType),
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    pub(crate) fn get_del(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => {
                    self.delete(&vec![key.to_string()]);
                    Response::String(value.to_string())
                }
                _ => Response::Err(IncompatibleDataType),
            }
        } else {
            Response::Err(UnknownKey)
        }
    }

    pub(crate) fn set(&self, key: &String, value: &String) -> Response {
        self.data
            .insert(key.to_string(), DbValue::String(value.to_string()));
        Response::String(String::from(OK))
    }

    pub(crate) fn incr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i64>() {
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

    pub(crate) fn incr_by(&self, key: &String, increment: &i64) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i64>() {
                    Ok(x) => {
                        let y = x + increment;
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

    pub(crate) fn decr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i64>() {
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

    pub(crate) fn decr_by(&self, key: &String, increment: &i64) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String(value) => match value.parse::<i64>() {
                    Ok(x) => {
                        let y = x - increment;
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

    fn get_string_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::String(_)))
    }
}