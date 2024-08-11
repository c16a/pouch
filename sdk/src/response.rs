use serde::{Deserialize, Serialize};
use std::collections::HashSet;

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Response {
    List { values: Vec<String> },
    Set { values: HashSet<String> },
    Err { error: Error },
    AffectedKeys { affected_keys: u64 },
    Count { count: u64 },
    StringValue { value: String },
    IntValue { value: i64 },
    BooleanValue { value: bool },
}

impl Response {
    pub fn to_json(&self) -> serde_json::Result<String> {
        serde_json::to_string(self)
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum Error {
    UnknownCommand,
    UnknownKey,
    IncompatibleDataType,
    NotInteger,
    TimeWentBackwards,
}
