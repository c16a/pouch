use crate::processor::db::InMemoryDb;
use crate::processor::spec::Processor;
use pouch_sdk::command::Command;
use std::fmt::Debug;
use std::fs::{File, OpenOptions};
use std::io::{BufRead, BufReader, Error, ErrorKind, Result as IOResult, Write};

pub(crate) struct WAL {
    file: File,
}

impl WAL {
    pub(crate) fn new(path: &str) -> IOResult<WAL> {
        let file = OpenOptions::new()
            .create(true)
            .append(true)
            .read(true)
            .open(path)?;
        Ok(WAL { file })
    }

    pub(crate) fn log(&mut self, cmd: &Command) -> IOResult<()> {
        let serialised = serde_json::to_string(cmd).unwrap();
        writeln!(self.file, "{}", serialised)?;
        self.file.flush().expect("failed to write WAL entry");
        Ok(())
    }

    pub(crate) fn replay(&self, db: &mut InMemoryDb) -> IOResult<usize> {
        let metadata = &self.file.metadata()?;
        if metadata.len() == 0 {
            return Err(Error::new(ErrorKind::InvalidData, "WAL file is empty"));
        }

        let reader = BufReader::new(&self.file);
        let mut count = 0;
        for line in reader.lines() {
            let line = line?;
            let cmd: Command = serde_json::from_str(&line).unwrap();
            db.cmd(cmd, None);
            count += 1;
        }
        Ok(count)
    }
}
