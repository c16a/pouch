use std::env;
use std::process::exit;
use tokio::io;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::TcpStream;

#[tokio::main]
async fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!("Usage: {} <server_address> <port>", args[0]);
        exit(1);
    }

    let server_address = format!("{}:{}", args[1], args[2]);

    let stream = TcpStream::connect(&server_address)
        .await
        .unwrap_or_else(|e| {
            eprintln!("Failed to connect: {}", e);
            exit(1);
        });
    println!("Connected to {}", server_address);

    handle_interactive_loop(stream).await;
}

async fn handle_interactive_loop(mut stream: TcpStream) {
    let (read_half, mut write_half) = stream.split();
    let mut reader = BufReader::new(read_half).lines();
    let stdin = io::stdin();
    let mut stdin_reader = BufReader::new(stdin).lines();

    let mut stdout = io::stdout();

    loop {
        stdout.write_all("pouch-cli> ".to_string().as_bytes()).await.unwrap();
        stdout.flush().await.unwrap();

        tokio::select! {
            Ok(Some(command)) = stdin_reader.next_line() => {
                if command.trim().is_empty() {
                    continue;
                }
                if command.to_lowercase() == "quit" {
                    println!("Exiting...");
                    break;
                }

                // Send the command to the server
                // Don't write a new line explicitly because the user would hit the ENTER key themselves.
                write_half.write_all(command.as_bytes()).await.expect("Failed to write to server");
                
                // Wait for the server response and print it back
                if let Ok(Some(response)) = reader.next_line().await {
                    stdout.write_all(format!("pouch-server> {}\n", response).as_bytes()).await.unwrap();
                    stdout.flush().await.unwrap();
                } else {
                    stdout.write_all(b"No response from server.\n").await.unwrap();
                    stdout.flush().await.unwrap();
                }
            }

            Ok(Some(response)) = reader.next_line() => {
                println!("server> {}", response);
                io::stdout().flush().await.expect("Failed to initialise prompt");
            }
        }
    }
}
