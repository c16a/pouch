use crate::command::Command;
use crate::response::Response;
use crate::wal::WAL;

pub(crate) trait Processor: Send + Sync {
    fn cmd(&self, cmd: Command, wal: Option<&mut WAL>) -> Response;
}