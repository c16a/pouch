use std::collections::HashSet;

#[derive(Debug, Clone, PartialEq)]
pub(crate) enum Response {
    SimpleString { value: String },
    List { values: Vec<String> },
    Set { values: HashSet<String> },
}

impl Response {
    pub(crate) fn to_vec(self) -> Vec<u8> {
        match self {
            Response::SimpleString { value } => {
                let v = "> ".to_owned() + &value + "\n";
                v.as_bytes().to_vec()
            }
            Response::List { values } => {
                let mut result = "> ".to_owned() + &values.join("\n> ");
                result.push('\n');
                result.into_bytes()
            }
            Response::Set { values } => {
                let mut result = "> ".to_owned() + &values.iter().cloned().collect::<Vec<String>>().join("\n> ");
                result.push('\n');
                result.into_bytes()
            }
        }
    }
}

pub(crate) const UNKNOWN_COMMAND: &str = "(error) unknown command";
pub(crate) const UNKNOWN_KEY: &str = "(error) unknown key";
pub(crate) const OK: &str = "OK";
pub(crate) const TRUE: &str = "true";
pub(crate) const FALSE: &str = "true";
pub(crate) const INCOMPATIBLE_DATA_TYPE: &str = "(error) incompatible data type";
