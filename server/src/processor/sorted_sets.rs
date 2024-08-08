use crate::processor::db::{DbValue, InMemoryDb};
use pouch_sdk::response::Error::{IncompatibleDataType, UnknownKey};
use pouch_sdk::response::Response;
use crate::structures::sorted_set::{SortedSet, SortedSetAddReturnType};
use dashmap::mapref::one::{Ref, RefMut};
use std::collections::HashMap;

impl InMemoryDb {
    fn get_sorted_set_ref(&self, key: &String) -> Option<Ref<String, DbValue>> {
        self.get_value_ref(key, |value| matches!(value, DbValue::SortedSet(_)))
    }

    fn get_sorted_set_ref_mut(&self, key: &String) -> Option<RefMut<String, DbValue>> {
        self.get_value_ref_mut(key, |value| matches!(value, DbValue::SortedSet(_)))
    }
    pub(crate) fn zadd(&self, key: &String, values: &HashMap<String, i64>) -> Response {
        if let Some(mut sorted_set_ref) = self.get_sorted_set_ref_mut(&key) {
            if let DbValue::SortedSet(sorted_set) = sorted_set_ref.value_mut() {
                let affected_rows =
                    sorted_set.add_all(values.to_owned(), Some(SortedSetAddReturnType::Added));
                Response::Integer(affected_rows)
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            let mut sorted_set = SortedSet::new();
            let affected_rows =
                sorted_set.add_all(values.to_owned(), Some(SortedSetAddReturnType::Added));
            self.data
                .insert(key.to_string(), DbValue::SortedSet(sorted_set));
            Response::Integer(affected_rows)
        }
    }

    pub(crate) fn zcard(&self, key: &String) -> Response {
        if let Some(sorted_set_ref) = self.get_sorted_set_ref(&key) {
            if let DbValue::SortedSet(sorted_set) = sorted_set_ref.value() {
                Response::Integer(sorted_set.cardinality() as i64)
            } else {
                Response::Err(IncompatibleDataType)
            }
        } else {
            Response::Err(UnknownKey)
        }
    }
}
