use tiny_http::{Response, Server};

fn main() {
    let server = Server::http("0.0.0.0:8080").expect("Failed to start server");
    println!("Server listening on http://localhost:8080");
    for request in server.incoming_requests() {
        let response = Response::from_string("Hello, World!\n");
        if let Err(e) = request.respond(response) {
            eprintln!("Failed to send response: {e}");
        }
    }
}
