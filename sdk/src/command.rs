use serde::{Deserialize, Serialize};
use serde_json::Result;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "action")]
pub enum Command {
    #[serde(rename = "GET")]
    Get {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "GETDEL")]
    GetDel {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "SET")]
    Set {
        #[serde(rename = "key")]
        key: String,

        #[serde(rename = "value")]
        value: String,

        #[serde(rename = "expiry_seconds")]
        expiry_seconds: u64,

        #[serde(rename = "expiry_ts")]
        #[serde(default)]
        expiry_ts: u64,
    },
    #[serde(rename = "DELETE")]
    Delete {
        #[serde(rename = "keys")]
        keys: Vec<String>,
    },
    #[serde(rename = "LPUSH")]
    LPush {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "values")]
        values: Vec<String>,
    },
    #[serde(rename = "RPUSH")]
    RPush {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "values")]
        values: Vec<String>,
    },
    #[serde(rename = "LRANGE")]
    LRange {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "start")]
        start: Option<usize>,
        #[serde(rename = "end")]
        end: Option<usize>,
    },
    #[serde(rename = "LLEN")]
    LLen {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "LPOP")]
    LPop {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "RPOP")]
    RPop {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "EXISTS")]
    Exists {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "INCR")]
    Incr {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "INCRBY")]
    IncrBy {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "increment")]
        increment: i64,
    },
    #[serde(rename = "DECR")]
    Decr {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "DECRBY")]
    DecrBy {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "decrement")]
        decrement: i64,
    },
    #[serde(rename = "SADD")]
    SAdd {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "values")]
        values: Vec<String>,
    },
    #[serde(rename = "SCARD")]
    SCard {
        #[serde(rename = "key")]
        key: String,
    },
    #[serde(rename = "SINTER")]
    SInter {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "others")]
        others: Vec<String>,
    },
    #[serde(rename = "SDIFF")]
    SDiff {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "others")]
        others: Vec<String>,
    },
    #[serde(rename = "ZADD")]
    ZAdd {
        #[serde(rename = "key")]
        key: String,
        #[serde(rename = "values")]
        values: HashMap<String, i64>,
    },
    #[serde(rename = "ZCARD")]
    ZCard {
        #[serde(rename = "key")]
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
        let json_str = r#"{"action":"GET","key":"mykey"}"#;
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
            expiry_seconds: 3600,
            expiry_ts: 0,
        };
        let json_str = cmd.to_json().unwrap();
        assert_eq!(
            json_str,
            r#"{"action":"SET","key":"mykey","value":"myvalue","expiry_seconds":3600,"expiry_ts":0}"#
        );
    }

    #[test]
    fn test_zadd_command_from_json() {
        let json_str = r#"{"action":"ZADD","key":"myset","values":{"key_1":1,"key_2":2}}"#;
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
            r#"{"action":"LRANGE","key":"mylist","start":0,"end":10}"#
        );
    }
}
