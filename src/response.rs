#[derive(Debug, Clone)]
pub(crate) enum Response {
    SimpleString { value: String },
    List { values: Vec<String> },
}

impl Response {
    pub(crate) fn to_vec(self) -> Vec<u8> {
        match self {
            Response::SimpleString { value } => {
                let v = "answer> ".to_owned() + &value + "\n";
                v.as_bytes().to_vec()
            }
            Response::List { values } => {
                let mut result = "answer> ".to_owned() + &values.join("\nanswer> ");
                result.push('\n');
                result.into_bytes()
            }
        }
    }
}