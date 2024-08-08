use std::collections::HashSet;

#[derive(Debug, Clone, PartialEq)]
pub(crate) enum Response {
    String(String),
    Integer(i64),
    List { values: Vec<String> },
    Set { values: HashSet<String> },
    Err(Error),
}

#[derive(Debug, Clone, PartialEq)]
pub(crate) enum Error {
    UnknownCommand,
    UnknownKey,
    IncompatibleDataType,
    NotInteger,
}

impl Error {
    fn to_string(&self) -> String {
        match self {
            Error::UnknownCommand => String::from("unknown command"),
            Error::UnknownKey => String::from("unknown key"),
            Error::IncompatibleDataType => {
                String::from("WRONGTYPE Operation against a key holding the wrong kind of value")
            }
            Error::NotInteger => String::from("value is not an integer or out of range"),
        }
    }
}

impl Response {
    pub(crate) fn to_vec(self) -> Vec<u8> {
        match self {
            Response::String(value) => {
                let v = format!("\"{}\"\n", value);
                v.as_bytes().to_vec()
            }
            Response::Integer(i) => {
                let v = format!("(integer) {}\n", i);
                v.as_bytes().to_vec()
            }
            Response::List { values } => {
                let mut result = "> ".to_owned() + &values.join("\n> ");
                result.push('\n');
                result.into_bytes()
            }
            Response::Set { values } => {
                let mut result =
                    "> ".to_owned() + &values.iter().cloned().collect::<Vec<String>>().join("\n> ");
                result.push('\n');
                result.into_bytes()
            }
            Response::Err(err) => {
                let v = format!("(error) {}\n", err.to_string());
                v.as_bytes().to_vec()
            }
        }
    }
}

pub(crate) const OK: &str = "OK";
