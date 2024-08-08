use std::collections::HashSet;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum Response {
    String(String),
    Integer(i64),
    List { values: Vec<String> },
    Set { values: HashSet<String> },
    Err(Error),
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
}

pub const OK: &str = "OK";
