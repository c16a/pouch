use crate::processor::db::{DbValue, InMemoryDb};
use pouch_sdk::response::Error::{IncompatibleDataType, UnknownKey};
use pouch_sdk::response::Response;
use dashmap::mapref::one::Ref;
use std::collections::HashSet;

impl InMemoryDb {
    pub(crate) fn sadd(&self, key: &String, values: &Vec<String>) -> Response {
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

    pub(crate) fn scard(&self, key: &String) -> Response {
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

    fn get_set_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::Set(_)))
    }

    pub(crate) fn sinter(&self, key: &String, others: &Vec<String>) -> Response {
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

    pub(crate) fn sdiff(&self, key: &String, others: &Vec<String>) -> Response {
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
}
