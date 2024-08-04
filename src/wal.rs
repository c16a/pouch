use std::fs::{File, OpenOptions};
use std::io::{BufRead, BufReader, Error, ErrorKind, Write, Result};

use tokio::io;

use crate::command::Command;
use crate::db::InMemoryDb;
use crate::processor::Processor;

pub(crate) struct WAL {
    file: File,
}

impl WAL {
    pub(crate) fn new(path: &str) -> Result<WAL> {
        let file = OpenOptions::new().create(true).append(true).read(true).open(path)?;
        Ok(WAL { file })
    }

    pub(crate) fn log(&mut self, cmd: &Command) -> io::Result<()> {
        let serialised = serde_json::to_string(cmd).unwrap();
        writeln!(self.file, "{}", serialised)?;
        self.file.flush().expect("failed to write WAL entry");
        Ok(())
    }

    pub(crate) fn replay(&self, db: &mut InMemoryDb) -> io::Result<usize> {
        let metadata = &self.file.metadata()?;
        if metadata.len() == 0 {
            return Err(Error::new(ErrorKind::InvalidData, "WAL file is empty"));
        }

        let reader = BufReader::new(&self.file);
        let mut count = 0;
        for line in reader.lines() {
            let line = line?;
            let cmd: Command = serde_json::from_str(&line).unwrap();
            match cmd {
                Command::Set { key, value } => {
                    db.set(&key, &value);
                }
                Command::Delete { key } => {
                    db.remove(&key);
                }
                Command::LPush { key, value } => {
                    db.lpush(&key, &value);
                }
                Command::RPush { key, value } => {
                    db.rpush(&key, &value);
                }
                Command::Incr { key } => {
                    db.incr(&key);
                }
                Command::Decr { key } => {
                    db.decr(&key);
                }
                _ => {} // Ignore other commands during replay
            }
            count += 1;
        }
        Ok(count)
    }
}