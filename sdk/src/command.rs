use serde::{Deserialize, Serialize};
use serde_json::Result;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Command {
    Get {
        key: String,
    },
    GetDel {
        key: String,
    },
    Set {
        key: String,
        value: String,
    },
    Delete {
        keys: Vec<String>,
    },
    LPush {
        key: String,
        values: Vec<String>,
    },
    RPush {
        key: String,
        values: Vec<String>,
    },
    LRange {
        key: String,
        start: Option<usize>,
        end: Option<usize>,
    },
    LLen {
        key: String,
    },
    LPop {
        key: String,
    },
    RPop {
        key: String,
    },
    Exists {
        key: String,
    },
    Incr {
        key: String,
    },
    IncrBy {
        key: String,
        increment: i64,
    },
    Decr {
        key: String,
    },
    DecrBy {
        key: String,
        decrement: i64,
    },
    SAdd {
        key: String,
        values: Vec<String>,
    },
    SCard {
        key: String,
    },
    SInter {
        key: String,
        others: Vec<String>,
    },
    SDiff {
        key: String,
        others: Vec<String>,
    },
    ZAdd {
        key: String,
        values: HashMap<String, i64>,
    },
    ZCard {
        key: String,
    },
}

impl Command {
    pub fn from_json(json_str: &str) -> Result<Command> {
        serde_json::from_str(json_str)
    }

    pub(crate) fn to_json(&self) -> Result<String> {
        serde_json::to_string(self)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_command_from_json() {
        let json_str = r#"{"Get":{"key":"mykey"}}"#;
        let cmd = Command::from_json(json_str).unwrap();
        match cmd {
            Command::Get { key } => assert_eq!(key, "mykey"),
            _ => panic!("Expected Command::Get"),
        }
    }

    #[test]
    fn test_set_command_to_json() {
        let cmd = Command::Set {
            key: "mykey".to_string(),
            value: "myvalue".to_string(),
        };
        let json_str = cmd.to_json().unwrap();
        assert_eq!(json_str, r#"{"Set":{"key":"mykey","value":"myvalue"}}"#);
    }

    #[test]
    fn test_zadd_command_from_json() {
        let json_str = r#"{"ZAdd":{"key":"myset","values":{"key_1":1,"key_2":2}}}"#;
        let cmd = Command::from_json(json_str).unwrap();
        match cmd {
            Command::ZAdd { key, values } => {
                assert_eq!(key, "myset");
                assert_eq!(values.get("key_1"), Some(&1));
                assert_eq!(values.get("key_2"), Some(&2));
            }
            _ => panic!("Expected Command::ZAdd"),
        }
    }

    #[test]
    fn test_lrange_command_to_json() {
        let cmd = Command::LRange {
            key: "mylist".to_string(),
            start: Some(0),
            end: Some(10),
        };
        let json_str = cmd.to_json().unwrap();
        assert_eq!(
            json_str,
            r#"{"LRange":{"key":"mylist","start":0,"end":10}}"#
        );
    }
}
