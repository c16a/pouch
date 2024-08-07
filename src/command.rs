use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub(crate) enum Command {
    Get {
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
                    Some(Command::Get {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "SET" => {
                if parts.len() == 3 {
                    Some(Command::Set {
                        key: parts[1].to_string(),
                        value: parts[2].to_string(),
                    })
                } else {
                    None
                }
            }
            "DEL" => {
                if parts.len() >= 2 {
                    let mut keys = Vec::new();
                    parts[1..].iter().for_each(|s| keys.push(s.to_string()));
                    Some(Command::Delete { keys })
                } else {
                    None
                }
            }
            "LPUSH" => {
                if parts.len() >= 3 {
                    let mut values = Vec::new();
                    parts[2..].iter().for_each(|s| values.push(s.to_string()));
                    Some(Command::LPush {
                        key: parts[1].to_string(),
                        values,
                    })
                } else {
                    None
                }
            }
            "RPUSH" => {
                if parts.len() >= 3 {
                    let mut values = Vec::new();
                    parts[2..].iter().for_each(|s| values.push(s.to_string()));
                    Some(Command::RPush {
                        key: parts[1].to_string(),
                        values,
                    })
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
                            Err(_err) => None,
                        };

                        if parts.len() >= 4 {
                            end = match parts[3].parse() {
                                Ok(val) => Some(val),
                                Err(_err) => None,
                            }
                        }
                    }

                    Some(Command::LRange { key, start, end })
                } else {
                    None
                }
            }
            "LLEN" => {
                if parts.len() == 2 {
                    Some(Command::LLen {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "LPOP" => {
                if parts.len() == 2 {
                    Some(Command::LPop {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "RPOP" => {
                if parts.len() == 2 {
                    Some(Command::RPop {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "EXISTS" => {
                if parts.len() == 2 {
                    Some(Command::Exists {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "INCR" => {
                if parts.len() == 2 {
                    Some(Command::Incr {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "INCRBY" => {
                if parts.len() == 3 {
                    match parts[2].parse::<i64>() {
                        Ok(val) => Some(Command::IncrBy {
                            key: parts[1].to_string(),
                            increment: val,
                        }),
                        Err(_) => None,
                    }
                } else {
                    None
                }
            }
            "DECR" => {
                if parts.len() == 2 {
                    Some(Command::Decr {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "DECRBY" => {
                if parts.len() == 3 {
                    match parts[2].parse::<i64>() {
                        Ok(val) => Some(Command::DecrBy {
                            key: parts[1].to_string(),
                            decrement: val,
                        }),
                        Err(_) => None,
                    }
                } else {
                    None
                }
            }
            "SADD" => {
                if parts.len() >= 3 {
                    let mut values = Vec::new();
                    parts[2..].iter().for_each(|s| values.push(s.to_string()));
                    Some(Command::SAdd {
                        key: parts[1].to_string(),
                        values,
                    })
                } else {
                    None
                }
            }
            "SCARD" => {
                if parts.len() == 2 {
                    Some(Command::SCard {
                        key: parts[1].to_string(),
                    })
                } else {
                    None
                }
            }
            "SINTER" => {
                if parts.len() >= 3 {
                    let mut values = Vec::new();
                    parts[2..].iter().for_each(|s| values.push(s.to_string()));
                    Some(Command::SInter {
                        key: parts[1].to_string(),
                        others: values,
                    })
                } else {
                    None
                }
            }
            "SDIFF" => {
                if parts.len() >= 3 {
                    let mut values = Vec::new();
                    parts[2..].iter().for_each(|s| values.push(s.to_string()));
                    Some(Command::SDiff {
                        key: parts[1].to_string(),
                        others: values,
                    })
                } else {
                    None
                }
            }
            _ => None,
        }
    }
}
