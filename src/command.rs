use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub(crate) enum Command {
    Get { key: String },
    Set { key: String, value: String },
    Delete { key: String },
    LPush { key: String, value: String },
    RPush { key: String, value: String },
    LRange { key: String, start: Option<usize>, end: Option<usize> },
}

impl Command {
    pub(crate) fn from_slice(b: &[u8]) -> Option<Command> {
        if b.len() == 0 {
            return None;
        }

        let s = std::str::from_utf8(&b).ok()?;

        let parts: Vec<&str> = s.split_whitespace().collect();

        if parts.is_empty() {
            return None;
        }

        match parts[0] {
            "GET" => {
                if parts.len() == 2 {
                    Some(Command::Get { key: parts[1].to_string() })
                } else {
                    None
                }
            }
            "SET" => {
                if parts.len() == 3 {
                    Some(Command::Set { key: parts[1].to_string(), value: parts[2].to_string() })
                } else {
                    None
                }
            }
            "DELETE" => {
                if parts.len() == 2 {
                    Some(Command::Delete { key: parts[1].to_string() })
                } else {
                    None
                }
            }
            "LPUSH" => {
                if parts.len() == 3 {
                    Some(Command::LPush { key: parts[1].to_string(), value: parts[2].to_string() })
                } else {
                    None
                }
            }
            "RPUSH" => {
                if parts.len() == 3 {
                    Some(Command::RPush { key: parts[1].to_string(), value: parts[2].to_string() })
                } else {
                    None
                }
            }
            "LRANGE" => {
                if parts.len() >= 2 {
                    let key = parts[1].to_string();

                    let mut start: Option<usize> = None;
                    let mut end: Option<usize> = None;

                    if parts.len() >= 3 {
                        start = match parts[2].parse() {
                            Ok(val) => Some(val),
                            Err(_err) => None
                        };

                        if parts.len() >= 4 {
                            end = match parts[3].parse() {
                                Ok(val) => Some(val),
                                Err(_err) => None
                            }
                        }
                    }

                    Some(Command::LRange { key, start, end })
                } else {
                    None
                }
            }
            _ => None
            // EXISTS - checks for key
            // KEYS - checks for all keys that match a pattern
            // DBSIZE - checks for number of keys
        }
    }
}