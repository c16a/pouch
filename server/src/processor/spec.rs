use pouch_sdk::response::Response;
use crate::wal::WAL;
use pouch_sdk::command::Command;

pub(crate) trait Processor: Send + Sync {
    fn cmd(&self, cmd: Command, wal: Option<&mut WAL>) -> Response;
}
