# httpws

httpws is a lightweight and fast HTTP and WebSocket framework for Go. It provides a simple and flexible way to build web applications with middleware support, routing, and WebSocket communication.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
    - [Router](#router)
    - [Middleware](#middleware)
    - [Server Connection](#server-connection)
    - [WebSocket](#websocket)
- [Implementation Details](#implementation-details)
    - [HTTP Parsing](#http-parsing)
    - [WebSocket Protocol](#websocket-protocol)
    - [Module summary](#module-summary)

## Installation

To install httpws, use the `go get` command:

```sh
go get github.com/lvkeliang/httpws
```

## Quick Start
Here is a simple example of how to use the framework to create a web application that responds to different HTTP methods and handles WebSocket requests:
```Go
package main

import (
  "fmt"
  "github.com/lvkeliang/httpws/router"
  "github.com/lvkeliang/httpws/server"
  "log"
)

// Here is a simple example of how to use the framework to build a simple web application:
func main() {
  // Create a router
  r := router.NewRouter()
  r.HandleFunc("GET", "/", indexMiddleware)
  r.HandleFunc("POST", "/hello", loggingMiddleware, nameMiddleware, printFormData, helloMiddleware)
  r.HandleFunc("GET", "/ws", handleWebSocket)

  log.Println("Starting server on :8080")
  r.ListenAndServe(":8080")
}

// Responds to a request to the root path
func indexMiddleware(next router.HandlerFunc) router.HandlerFunc {
  return func(c server.Conn) {
    c.Message.Print()
    c.WriteResponse(200, "OK", []byte("Welcome to my website!"))
    next(c)
  }
}

// Logs the received request data before processing it
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        c.Message.Print()
        next(c)
    }
}

// Adds a middleware function that sets the name data
func nameMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        value, _ := c.Message.ReadFormData()
        c.Set("name", value["name"]) // Set the name data
        next(c)
    }
}

// Adds a middleware function that prints the form data
func printFormData(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        fmt.Println(c.Message.ReadFormData())
        next(c)
    }
}

// Adds a middleware function that responds with a greeting message and sets a cookie
func helloMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        name, ok := c.Get("name")
        if !ok {
            name = "World"
        }

        c.WriteResponse(200, "OK", []byte(fmt.Sprintf("Hello, %s!", name)),
            map[string]string{"set-cookie": fmt.Sprintf("name=%v; Max-Age=3600; Domain=localhost;Secure; Path=/; Version=1", name)})
        next(c)
    }
}

// handleWebSocket handles WebSocket requests
func handleWebSocket(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        // Handshake upgrade
        err := c.UpgradeToWebSocket()
        if err != nil {
            log.Println(err)
            return
        }
        // Loop to read and write messages
        for {
            // Read a message from the WebSocket connection and get its opcode, payload and error
            opCode, payload, err := c.ReadWebSocketMessage()
            if err != nil { // If there is an error, handle it and break the loop
                c.WebSocketHandleError(err)
                log.Println(err)
                break
            }
            // Print the received message to the console
            fmt.Printf("Received context: opCode = %d, payload = %s\n", opCode, string(payload))
            // Write the received message back to the WebSocket connection, if there is an error, handle it and break the loop
            if err := c.WriteWebSocketMessage(opCode, payload); err != nil {
                c.WebSocketHandleError(err)
                break
            }
        }
    }
}

```

## Usage

### Router

The router is responsible for matching incoming requests to registered handlers based on the HTTP method and path. To create a new router, use the `NewRouter` function:

```Go
r := router.NewRouter()
```

To register a handler for a specific method and path, use the `HandleFunc` method of the router. The handler is a function that takes a `server.Conn` as an argument and performs some action on it. You can also pass one or more middleware functions as arguments to `HandleFunc`, which will be executed before the handler. A middleware function is a function that takes a handler function as an argument and returns a new handler function.

```Go
r.HandleFunc("GET", "/", indexMiddleware)
r.HandleFunc("POST", "/hello", loggingMiddleware, nameMiddleware, printFormData, helloMiddleware)
```

To start the server and listen for incoming requests, use the `ListenAndServe` method of the router. You can pass a port number as an argument.

```Go
r.ListenAndServe(":8080")
```

### Middleware

Middleware is a way to add additional functionality to the request processing pipeline. A middleware function is a function that takes a handler function as an argument and returns a new handler function. The middleware function can perform some action before or after calling the next handler function, or modify the server connection or the request message.

For example, here is a middleware function that logs the received request data before processing it:

```Go
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        c.Message.Print()
        next(c)
    }
}
```

Here is another middleware function that sets the name data from the form and passes it to the next handler:

```Go
func nameMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        value, _ := c.Message.ReadFormData()
        c.Set("name", value["name"]) // Set the name data
        next(c)
    }
}
```

You can register one or more middleware functions for a specific method and path using the `HandleFunc` method of the router. The middleware functions will be executed in the order they are passed, and the final handler will be passed to the last middleware function.

```Go
r.HandleFunc("POST", "/hello", loggingMiddleware, nameMiddleware, printFormData, helloMiddleware)
```

### Server Connection

The server connection is an abstraction of the TCP connection between the server and the client. It provides methods to read and write HTTP and WebSocket messages, as well as store and retrieve data associated with the connection.

The server connection is passed as an argument to the handler and middleware functions. You can access its fields and methods to perform various actions on it.

For example, you can use the `Message` field to access the request message, which is an instance of `server.Message`. The `server.Message type` provides methods to parse and manipulate HTTP messages, such as reading headers, body, form data, etc.

```Go
c.Message.Print() // Print the request message
value, _ := c.Message.ReadFormData() // Read the form data from the request body
```

You can also use the `WriteResponse` method to write an HTTP response to the client. You can pass arguments such as status code, reason phrase, response body, and response headers.

```Go
c.WriteResponse(200, "OK", []byte("Hello, World!")) // Write a simple response with status code 200 and body "Hello, World!"
c.WriteResponse(200, "OK", []byte(fmt.Sprintf("Hello, %s!", name)),
    map[string]string{"set-cookie": fmt.Sprintf("name=%v; Max-Age=3600; Domain=localhost;Secure; Path=/; Version=1", name)}) // Write a response with status code 200 and body "Hello, name!", and set a cookie named "name" with value "name"
```

You can also use the `Set` and `Get` methods to store and retrieve data associated with the connection. This can be useful for passing data between middleware functions or handlers.

```Go
c.Set("name", value["name"]) // Set the name data
name, ok := c.Get("name") // Get the name data
```

### WebSocket

WebSocket is a protocol that allows bidirectional communication between the server and the client over a single TCP connection. It is useful for applications that require real-time updates or interactive features.

To handle WebSocket requests with httpws, you need to upgrade the server connection to a WebSocket connection using the `UpgradeToWebSocket` method. This method will perform a handshake with the client and return an error if it fails.

```Go
err := c.UpgradeToWebSocket()
    if err != nil {
        log.Println(err)
    return
}
```

After upgrading to WebSocket, you can use the `ReadWebSocketMessage` and `WriteWebSocketMessage` methods to read and write WebSocket messages. A WebSocket message consists of an opcode and a payload. The opcode indicates the type of message (such as text, binary, close, ping or pong), and the payload is a slice of bytes that contains the message data.

```Go
// Read a message from the WebSocket connection and get its opcode, payload and error
opCode, payload, err := c.ReadWebSocketMessage()
if err != nil { // If there is an error, handle it and break the loop
    c.WebSocketHandleError(err)
    log.Println(err)
    break
}

// Write a message to the WebSocket connection with the same opcode and payload, if there is an error, handle it and break the loop

if err := c.WriteWebSocketMessage(opCode, payload); err != nil {
    c.WebSocketHandleError(err)
    break
}
```

## Implementation Details

### HTTP Parsing

The framework uses a custom HTTP parser to parse incoming HTTP messages. The parser is a state machine that reads bytes from the TCP connection and updates its state according to the bytes. The parser can handle different types of HTTP messages, such as requests, responses, chunked encoding, multipart form data, etc.

The parser is designed to be fast and efficient, using minimal memory allocation and copying. The parser also supports pipelining, which means it can handle multiple HTTP messages on the same connection without waiting for the previous message to finish.

The parser exposes a `ReadFormData` method on the `server.Message` instances, which is used to read information from forms and returns it as a map. The `server.Message` type provides methods to access and manipulate the components of the HTTP message, such as headers, body, form data, etc.
### WebSocket Protocol

The framework implements the WebSocket protocol defined in RFC 6455. The protocol allows the server and the client to exchange messages over a single TCP connection, using a binary frame layer to encode the message type and length.

The framework provides methods to upgrade a server connection to a WebSocket connection, and to read and write WebSocket messages. The framework also handles WebSocket control frames, such as close, ping and pong, and performs the necessary masking and unmasking operations on the payload data.

The framework exposes some methods, such as `UpgradeToWebSocket`, `ReadWebSocketMessage` and `WriteWebSocketMessage`, that take a `server.Conn` as an argument and perform WebSocket operations on it. The framework also provides a helper method `WebSocketHandleError`, which is used to handle common WebSocket errors, such as sending a close frame or closing the connection.

### Module summary

This framework is used to implement simple HTTP and WebSocket services. It mainly consists of three modules: router, context and server. The following is a summary of these three modules:

<br/>

- The `router` module is used to define and manage routing rules and handler functions. It defines a Router struct, which contains a map type rules field, which is used to store the mapping relationship between routing rules and handler functions. It also defines a HandlerFunc type, which is a function type that accepts a server.Conn type as an argument and performs some operations on it. It also defines a Middleware type, which is also a function type that accepts a HandlerFunc type as an argument and returns a new HandlerFunc type value. Middleware functions can perform some operations before or after calling the next handler function, or modify the server connection or request message.

- The `router` module provides the following methods and functions:

  - The `NewRouter` function, which is used to create and return a new Router instance, and initialize its rules field to an empty map.
  - The `HandleFunc` method, which is used to add new routing rules. It accepts a method string, a pattern string and a set of middleware functions as arguments. It generates a key based on the method string and pattern string, and stores it with the middleware function combined by the Chain function as the value in the rules field.
  - The `Chain` function, which is used to combine multiple middleware functions together. It accepts a set of middleware functions as arguments and returns a new middleware function. When calling this new middleware function, it calls all the passed-in middleware functions in turn, and passes the final handler to the last middleware function.
  - The `ListenAndServe` method, which is used to start the server and listen for TCP connections on the specified address. When receiving a new connection, it creates and initializes a server.Conn instance, and calls the Serve method to handle this connection.
  - The `Serve` method, which is used to handle client connections. It looks up the corresponding handler function from the rules field according to the request method and path, and calls it to handle the request. If no matching handler function is found, it returns a 404 Not Found response.

<br/>

- The `context` module is used to parse and print HTTP messages. It defines a Context struct, which contains three fields: StartLine, Headers and Body, which are used to store the start line, header fields and message body of the message.

- The `context` module provides the following methods and functions:

  - The `NewContext` function, which is used to create a Context instance from the Req variable and return it. This function parses out the various parts of the message from Req and stores them in the Context instance.
  - The `Print` method, which is used to print the various parts of the Context instance for debugging purposes. This method uses functions provided by the fmt package to format output of start line, header fields and message body.
  - The `ReadFormData` method, which is used to read form-data from the message body Body and return a map type result and an error value. This method gets the content type (Content-Type) from the header fields Headers and parses out the boundary value. Then it converts the message body Body into a byte slice and uses the boundary as a delimiter to split it into multiple byte slices. Each byte slice represents a form data, containing two parts: header fields and value. Then it calls the parseHeader function to parse header fields, get name and value, and store them in map.
  - The `parseHeader` function, which is used to parse header fields, get name and value, and return them with an error value. This function converts header fields into strings and splits them into multiple parts by semicolons (;). Then it traverses each part, looking for parts that start with name= or filename=, and gets their values. If there is a file name, it uses file name as part of value and removes leading and trailing carriage return line feed characters (CRLF).

<br/>

- The `server` module is used to implement HTTP and WebSocket services. It defines a Conn struct, which contains five fields: Conn, Message, Data, mu and ws. The Conn field is a net.Conn type value, representing the TCP connection between the server and the client. The Message field is a pointer to a context.Context type, representing the parsed HTTP message. The Data field is a map type value, used to store and retrieve data related to the connection. The mu field is a sync.RWMutex type value, used to read and write lock the connection. The ws field is a bool type value, indicating whether it has been upgraded to a WebSocket connection.

- The `server` module provides the following methods and functions:

  - The `Set` method, which is used to set values across middleware. It accepts a string type key and an empty interface type value as arguments. It stores the key-value pair in the Data field and uses the mu field to write lock the connection.
  - The `Get` method, which is used to get values across middleware. It accepts a string type key as an argument and returns an empty interface type value and a bool type value as results. It retrieves the value corresponding to the key from the Data field and uses the mu field to read lock the connection.
  - The `WriteResponse` method, which is used to write a custom HTTP response to Conn. It accepts four arguments: statusCode, statusText, body and headers. statusCode is an integer type value, indicating the status code of the response. statusText is a string type value, indicating the reason phrase of the response. body is a byte slice type value, indicating the body of the response. headers is a variadic parameter, indicating the header fields of the response. This method uses the mu field to write lock the connection and returns an error value.
  - The `detectContentType` function, which is used to automatically detect the MIME type based on the content of body and return a string type result. This function determines whether body is HTML or JSON type based on its first byte, otherwise it uses binary stream type.
  - The `IsWebSocket` method, which is used to return whether Conn has been upgraded to a WebSocket connection. It gets the value corresponding to the “websocket” key from the Data field and returns it. If the Data field is empty, it creates an empty map first.
  - The `UpgradeToWebSocket` method, which is used to upgrade a Conn to a WebSocket connection by performing a handshake. It uses the mu field to write lock the connection and returns an error value. This method determines whether the request message meets the requirements of the WebSocket protocol based on its header fields and generates the Sec-WebSocket-Accept header in the response message. Then it writes the response message to Conn and sets the “websocket” key in Data field to true, indicating that it has been upgraded to a WebSocket connection.
  - The `ReadWebSocketMessage` method, which is used to read a message from a WebSocket connection and return its opcode, payload and error value. This method uses the mu field to read lock the connection. This method creates a bufio.Reader type buffered reader and uses readWebSocketFrame function to read frames until it encounters the last frame. If opcode is close frame, it returns opcode, empty payload and EOF error. If opcode is ping frame, it sends a pong frame back and continues looping. If opcode is pong frame, it ignores it and continues looping.
  - The `readWebSocketFrame` function, which is used to read a frame from a WebSocket connection and return its fin bit, opcode and payload and error value. This function reads first and second bytes from buffered reader and gets fin bit, opcode, MASK bit and payload length values. If payload length is 126 or 127, it means there are extended lengths behind it, so it needs to read two or eight more bytes from reader and merge them into extended length value. If payload length exceeds limit, it returns error value. Then it determines whether MASK bit is true, if so, it means there are four bytes of mask behind it, so it needs to read four more bytes from reader and store them in array. Then it creates slice payload for storing payload. It reads payload length data from reader and stores them in slice. If error occurs, it returns error value. Finally it determines whether MASK bit is true, if so, it means payload needs XOR operation. It traverses each byte of payload and XORs with corresponding byte of mask. Then it returns fin bit, opcode, payload and nil error value.
  - The `WriteWebSocketMessage` method, which is used to write a message to WebSocket connection. It accepts two arguments: opCode and payload. opCode is an integer type value, indicating message’s opcode. payload is a byte slice type value, indicating message’s payload. It uses mu field to write lock connection and returns an error value. This method creates a bytes.Buffer type buffer for storing WebSocket frame. Then it sets first byte of frame, containing fin bit and opcode. Assuming message is not fragmented, so fin bit is 1. Then it sets second byte of frame, containing mask bit and payload length. Not using mask, so mask bit is 0. According to size of payload length, it chooses appropriate encoding method and writes length field and extended length (if any) to buffer. Then it writes payload, not performing mask operation. Finally it writes buffer’s content to network connection and returns error value (if any).
  - The `CloseWebSocket` method, which is used to close WebSocket connection. It uses mu field to write lock connection and returns an error value. This method determines whether Conn has been upgraded to a WebSocket connection, if not, it returns error. Then it sends a close frame to the other side and waits for the other side to reply with a close frame. Then it closes underlying net.Conn and sets “websocket” key in Data field to false, indicating that WebSocket connection has been closed.
  - The `WebSocketHandleError` function, which is used to handle errors that occur when reading or writing WebSocket messages. It accepts an error type argument err and has no return value. This function determines whether error is EOF, if so, it means the other side closed connection and prints corresponding information. Then it determines whether error is a network error and whether it is a timeout error, if so, it means connection timed out and prints corresponding information. In other cases, it means an unexpected error occurred and prints corresponding information. Finally it calls CloseWebSocket method to close connection.

<br/>

The above code framework implements the basic functions of HTTP and WebSocket, but there are some places that can be improved or extended:

- For HTTP service, more header fields and body types can be added to support different content formats and encoding methods, such as JSON, XML, multipart/form-data, etc.
- For HTTP service, more status codes and reason phrases can be added to respond to different situations, such as 200 OK, 301 Moved Permanently, 400 Bad Request, 500 Internal Server Error, etc.
- For HTTP service, more middleware functions can be added to implement some common or specific functions, such as logging, authentication, compression, caching, etc.
- For WebSocket service, support for fragmented messages can be added, that is, splitting a large message into multiple frames and sending them, and recombining them at the receiving end.
- For WebSocket service, support for control frames can be added, that is, sending or receiving close frames, ping frames or pong frames while sending or receiving data frames.
- For WebSocket service, different processing methods for text and binary frames can be added, such as encoding or decoding text frames, compressing or decompressing binary frames, etc.
- For WebSocket service, more error handling and exception handling can be added, such as responding or reconnecting appropriately to situations such as connection interruption, timeout, format error, etc.

For users of this framework, the following points need to be considered:

- Users need to understand the basic principles and specifications of HTTP and WebSocket protocols in order to correctly construct request and response messages and handle different opcodes and payloads.
- Users need to use the methods and functions provided by the router module to define routing rules and handler functions and use the methods and functions provided by the context module to parse and print HTTP messages.
- Users need to use the methods and functions provided by the server module to set and get data related to the connection and upgrade or close WebSocket connections as needed and read or write WebSocket messages.