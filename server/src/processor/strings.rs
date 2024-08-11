use crate::processor::db::{DbValue, InMemoryDb};
use dashmap::mapref::one::Ref;
use pouch_sdk::response::Error::{IncompatibleDataType, NotInteger, TimeWentBackwards, UnknownKey};
use pouch_sdk::response::Response;
use std::time::{SystemTime, UNIX_EPOCH};

impl InMemoryDb {
    pub(crate) fn get(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String { value, expiry_ts } => {
                    match SystemTime::now().duration_since(UNIX_EPOCH) {
                        Ok(now) => {
                            if now.as_secs() > *expiry_ts {
                                self.data.remove(&value.to_string());
                                Response::Err { error: UnknownKey }
                            } else {
                                Response::StringValue {
                                    value: value.to_string(),
                                }
                            }
                        }
                        Err(_err) => Response::Err {
                            error: TimeWentBackwards,
                        },
                    }
                }
                _ => Response::Err {
                    error: IncompatibleDataType,
                },
            }
        } else {
            Response::Err { error: UnknownKey }
        }
    }

    pub(crate) fn get_del(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String { value, expiry_ts } => {
                    self.data.remove(&value.to_string());
                    match SystemTime::now().duration_since(UNIX_EPOCH) {
                        Ok(now) => {
                            if now.as_secs() > *expiry_ts {
                                Response::Err { error: UnknownKey }
                            } else {
                                Response::StringValue {
                                    value: value.to_string(),
                                }
                            }
                        }
                        Err(_err) => Response::Err {
                            error: TimeWentBackwards,
                        },
                    }
                }
                _ => Response::Err {
                    error: IncompatibleDataType,
                },
            }
        } else {
            Response::Err { error: UnknownKey }
        }
    }

    pub(crate) fn set(&self, key: &String, value: &String, expiry_seconds: &u64) -> Response {
        match SystemTime::now().duration_since(UNIX_EPOCH) {
            Ok(now) => {
                let expiry_ts = now.as_secs() + expiry_seconds;
                self.data.insert(
                    key.to_string(),
                    DbValue::String {
                        value: value.to_string(),
                        expiry_ts,
                    },
                );
                Response::AffectedKeys { affected_keys: 1 }
            }
            Err(_err) => Response::Err {
                error: TimeWentBackwards,
            },
        }
    }

    pub(crate) fn incr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String { value, expiry_ts } => match value.parse::<i64>() {
                    Ok(x) => {
                        let y = x + 1;
                        self.data.insert(
                            key.to_string(),
                            DbValue::String {
                                value: y.to_string(),
                                expiry_ts: *expiry_ts,
                            },
                        );
                        Response::IntValue { value: y }
                    }
                    Err(_err) => Response::Err { error: NotInteger },
                },
                _ => Response::Err {
                    error: IncompatibleDataType,
                },
            }
        } else {
            Response::Err { error: UnknownKey }
        }
    }

    pub(crate) fn incr_by(&self, key: &String, increment: &i64) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String { value, expiry_ts } => match value.parse::<i64>() {
                    Ok(x) => {
                        let y = x + increment;
                        self.data.insert(
                            key.to_string(),
                            DbValue::String {
                                value: y.to_string(),
                                expiry_ts: *expiry_ts,
                            },
                        );
                        Response::IntValue { value: y }
                    }
                    Err(_err) => Response::Err { error: NotInteger },
                },
                _ => Response::Err {
                    error: IncompatibleDataType,
                },
            }
        } else {
            Response::Err { error: UnknownKey }
        }
    }

    pub(crate) fn decr(&self, key: &String) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String { value, expiry_ts } => match value.parse::<i64>() {
                    Ok(x) => {
                        let y = x - 1;
                        self.data.insert(
                            key.to_string(),
                            DbValue::String {
                                value: y.to_string(),
                                expiry_ts: *expiry_ts,
                            },
                        );
                        Response::IntValue { value: y }
                    }
                    Err(_err) => Response::Err { error: NotInteger },
                },
                _ => Response::Err {
                    error: IncompatibleDataType,
                },
            }
        } else {
            Response::Err { error: UnknownKey }
        }
    }

    pub(crate) fn decr_by(&self, key: &String, increment: &i64) -> Response {
        if let Some(db_value) = self.data.get(key) {
            match db_value.value() {
                DbValue::String { value, expiry_ts } => match value.parse::<i64>() {
                    Ok(x) => {
                        let y = x - increment;
                        self.data.insert(
                            key.to_string(),
                            DbValue::String {
                                value: y.to_string(),
                                expiry_ts: *expiry_ts,
                            },
                        );
                        Response::IntValue { value: y }
                    }
                    Err(_err) => Response::Err { error: NotInteger },
                },
                _ => Response::Err {
                    error: IncompatibleDataType,
                },
            }
        } else {
            Response::Err { error: UnknownKey }
        }
    }

    fn get_string_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::String { .. }))
    }
}
