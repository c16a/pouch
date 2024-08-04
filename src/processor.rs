use crate::response::Response;

pub(crate) trait Processor: Send + Sync {
    fn get(&self, key: &String) -> Response;
    fn exists(&self, key: &String) -> Response;
    fn set(&self, key: &String, value: &String) -> Response;
    fn remove(&self, key: &String) -> Response;
    fn lpush(&self, key: &String, value: &String) -> Response;
    fn rpush(&self, key: &String, value: &String) -> Response;
    fn lrange(&self, key: &String, start: Option<usize>, end: Option<usize>) -> Response;
    fn incr(&self, key: &String) -> Response;
    fn decr(&self, key: &String) -> Response;
}