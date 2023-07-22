# httpws

httpws是一个轻量级且快速的Go语言HTTP和WebSocket框架。它提供了一种简单灵活的方式来构建具有中间件支持、路由和WebSocket通信的web应用程序。

## 目录

- [安装](#安装)
- [快速上手](#快速上手)
- [用法](#用法)
    - [路由](#路由)
    - [中间件](#中间件)
    - [服务器连接](#服务器连接)
    - [WebSocket](#websocket)
- [实现细节](#实现细节)
    - [HTTP解析](#HTTP解析)
    - [WebSocket协议](#WebSocket协议)
    - [模块总结](#模块总结)

## 安装

要安装 httpws，请使用`go get`命令：

```sh
go get github.com/lvkeliang/httpws
```

## 快速上手

这里有一个简单的例子，展示了如何使用这个框架来创建一个 web 应用，它可以响应不同的 HTTP 方法，并处理 WebSocket 请求：

```Go
package main

import (
	"fmt"
	"github.com/lvkeliang/httpws/router"
	"github.com/lvkeliang/httpws/server"
	"log"
)

// 下面是一个简单的使用示例，它展示了如何使用该框架来构建一个简单的 Web 应用：
func main() {
	// 创建一个路由
	r := router.NewRouter()
	r.HandleFunc("GET", "/", indexMiddleware)
	r.HandleFunc("POST", "/hello", loggingMiddleware, nameMiddleware, printFormData, helloMiddleware)
	r.HandleFunc("GET", "/ws", handleWebSocket)

	log.Println("Starting server on :8080")
	r.ListenAndServe(":8080")
}

// 用于回复一个访问根目录的消息
func indexMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		c.Message.Print()
		c.WriteResponse(200, "OK", []byte("Welcome to my website!"))
		next(c)
	}
}

// 用于在处理请求之前打印一条日志消息，记录收到的请求数据。
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		c.Message.Print()
		next(c)
	}
}

// 添加一个中间件函数，用于设置 name 数据
func nameMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		value, _ := c.Message.ReadFormData()
		c.Set("name", value["name"]) // 设置 name 数据
		next(c)
	}
}

// 添加一个中间件函数，用于打印表单
func printFormData(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		fmt.Println(c.Message.ReadFormData())
		next(c)
	}
}

// 添加一个中间件函数，用于回复打招呼消息，以及设置Cookie
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

// handleWebSocket 处理WebSocket请求
func handleWebSocket(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		// 握手升级
		err := c.UpgradeToWebSocket()
		if err != nil {
			log.Println(err)
			return
		}
		// 循环读取和写入消息
		for {
			// 从WebSocket连接中读取一个消息，并获取它的操作码、有效载荷和错误
			opCode, payload, err := c.ReadWebSocketMessage()
			if err != nil { // 如果出错，处理错误并退出循环
				c.WebSocketHandleError(err)
				log.Println(err)
				break
			}
			// 将收到的消息打印到控制台
			fmt.Printf("Received context: opCode = %d, payload = %s\n", opCode, string(payload))
			// 将收到的消息原样写回到WebSocket连接中，如果出错，处理错误并退出循环
			if err := c.WriteWebSocketMessage(opCode, payload); err != nil {
				c.WebSocketHandleError(err)
				break
			}
		}
	}
}
```


## 用法

### 路由

路由器负责根据 HTTP 方法和路径，将传入的请求匹配到已注册的处理器上。要创建一个新的路由器，使用`NewRouter`函数：

```Go
r := router.NewRouter()
```
要为特定的方法和路径注册一个处理器，请使用路由器的`HandleFunc`方法。处理器是一个函数，它接受一个`server.Conn`作为参数，并对它执行一些操作。你也可以将一个或多个中间件函数作为参数传递给`HandleFunc`，它们将在处理器之前执行。中间件函数是一个函数，它接受一个处理器函数作为参数，并返回一个新的处理器函数。


```Go
r.HandleFunc("GET", "/", indexMiddleware)
r.HandleFunc("POST", "/hello", loggingMiddleware, nameMiddleware, printFormData, helloMiddleware)
```
要启动服务器并监听传入的请求，使用路由器的`ListenAndServe`方法。你可以传递一个端口号作为参数。

```Go
r.ListenAndServe(":8080")
```

### 中间件

中间件是一种在请求处理流程中添加额外功能的方法。中间件函数是一个函数，它接受一个处理器函数作为参数，并返回一个新的处理器函数。中间件函数可以在调用下一个处理器函数之前或之后执行一些操作，或者修改服务器连接或请求消息。

例如，这里有一个中间件函数，它在处理请求之前记录收到的请求数据：

```Go
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        c.Message.Print()
        next(c)
    }
}
```
这里是另一个中间件函数，它从表单中获取 name 数据，并传递给下一个处理器：

```Go
func nameMiddleware(next router.HandlerFunc) router.HandlerFunc {
    return func(c server.Conn) {
        value, _ := c.Message.ReadFormData()
        c.Set("name", value["name"]) // 设置 name 数据
        next(c)
    }
}
```

你可以使用路由器的`HandleFunc`方法为特定的方法和路径注册一个或多个中间件函数。中间件函数将按照传递的顺序执行，最后一个处理器将传递给最后一个中间件函数。

```Go
r.HandleFunc("POST", "/hello", loggingMiddleware, nameMiddleware, printFormData, helloMiddleware)
```

### 服务器连接

服务器连接是服务器和客户端之间的 TCP 连接的抽象。它提供了读写 HTTP 和 WebSocket 消息，以及存储和检索与连接相关的数据的方法。

服务器连接作为参数传递给处理器和中间件函数。你可以访问它的字段和方法来对它执行各种操作。

例如，你可以使用`Message`字段来访问请求消息，它是`server.Message`的一个实例。`server.Message`类型提供了解析和操作 HTTP 消息的方法，比如读取头部、正文、表单数据等。

```Go
c.Message.Print() // 打印请求消息
value, _ := c.Message.ReadFormData() // 从请求正文中读取表单数据
```

你也可以使用`WriteResponse`方法来向客户端写入一个 HTTP 响应。你可以传递参数，如状态码、原因短语、响应正文和响应头部。

```Go
c.WriteResponse(200, "OK", []byte("Hello, World!")) // 写入一个简单的响应，状态码为 200，正文为 "Hello, World!"
c.WriteResponse(200, "OK", []byte(fmt.Sprintf("Hello, %s!", name)),
    map[string]string{"set-cookie": fmt.Sprintf("name=%v; Max-Age=3600; Domain=localhost;Secure; Path=/; Version=1", name)}) // 写入一个响应，状态码为 200，正文为 "Hello, name!"，并设置一个名为 "name" 的 cookie，值为 "name"
```

你也可以使用`Set`和`Get`方法来存储和检索与连接相关的数据。这对于在中间件函数或处理器之间传递数据很有用。

```Go
c.Set("name", value["name"]) // 设置 name 数据
name, ok := c.Get("name") // 获取 name 数据
```

### WebSocket

WebSocket 是一种协议，它允许服务器和客户端在单个 TCP 连接上进行双向通信。它适用于需要实时更新或交互功能的应用程序。

要使用 httpws 处理 WebSocket 请求，你需要使用`UpgradeToWebSocket`方法将服务器连接升级为 WebSocket 连接。这个方法将与客户端进行握手，并在失败时返回一个错误。

```Go
err := c.UpgradeToWebSocket()
    if err != nil {
        log.Println(err)
    return
}
```

升级为 WebSocket 后，你可以使用`ReadWebSocketMessage`和`WriteWebSocketMessage`方法来读写 WebSocket 消息。一个 WebSocket 消息由一个操作码和一个有效载荷组成。操作码表示消息的类型（如文本、二进制、关闭、ping 或 pong），有效载荷是一个字节切片，包含消息数据。

```Go
// 从 WebSocket 连接中读取一个消息，并获取它的操作码、有效载荷和错误
opCode, payload, err := c.ReadWebSocketMessage()
if err != nil { // 如果有错误，处理它并跳出循环
    c.WebSocketHandleError(err)
    log.Println(err)
    break
}

// 向 WebSocket 连接写入一个消息，使用相同的操作码和有效载荷，如果有错误，处理它并跳出循环

if err := c.WriteWebSocketMessage(opCode, payload); err != nil {
    c.WebSocketHandleError(err)
    break
}
```

## 实现细节

### HTTP解析

该框架使用一个自定义的 HTTP 解析器来解析传入的 HTTP 消息。解析器是一个状态机，它从 TCP 连接中读取字节，并根据字节更新其状态。解析器可以处理不同类型的 HTTP 消息，如请求、响应、分块编码、多部分表单数据等。

解析器的设计旨在快速和高效，使用最少的内存分配和复制。解析器还支持流水线，这意味着它可以在同一个连接上处理多个 HTTP 消息，而不需要等待前一个消息完成。

解析器对`server.Message`实例暴露了一个`ReadFormData`方法，它用于读取表单中的信息，并由map返回。`server.Message`类型提供了访问和操作 HTTP 消息组件的方法，如头部、正文、表单数据等。

### WebSocket协议

该框架实现了 RFC 6455 中定义的 WebSocket 协议。该协议允许服务器和客户端在单个 TCP 连接上交换消息，使用一个二进制帧层来编码消息类型和长度。

该框架提供了将服务器连接升级为 WebSocket 连接，以及读写 WebSocket 消息的方法。该框架还处理了 WebSocket 控制帧，如关闭、ping 和 pong，并对有效载荷数据进行必要的掩码和去掩码操作。

该框架暴露了一些方法，如`UpgradeToWebSocket`、`ReadWebSocketMessage`和`WriteWebSocketMessage`，它们接受一个`server.Conn`作为参数，并对其执行 WebSocket 操作。该框架还提供了一个辅助方法`WebSocketHandleError`，用于处理常见的 WebSocket 错误，如发送关闭帧或关闭连接。

### 模块总结

本框架用于实现简单的 HTTP 和 WebSocket 的服务。它主要包含三个模块：router、context 和 server。下面分别对这三个模块进行总结：

<br/>

- `router` 模块用于定义和管理路由规则和处理器函数。它定义了一个 Router 结构体，它包含了一个 map 类型的 rules 字段，用于存储路由规则和处理器函数的映射关系。它还定义了一个 HandlerFunc 类型，它是一个函数类型，接受一个 server.Conn 类型的参数，并对其执行一些操作。它还定义了一个 Middleware 类型，它也是一个函数类型，接受一个 HandlerFunc 类型的参数，并返回一个新的 HandlerFunc 类型的值。中间件函数可以在调用下一个处理器函数之前或之后执行一些操作，或者修改服务器连接或请求消息。

- `router` 模块提供了以下几个方法和函数：

    - `NewRouter` 函数，用于创建并返回一个新的 Router 实例，并初始化其 rules 字段为空的 map。
    - `HandleFunc` 方法，用于添加新的路由规则。它接受一个方法字符串、一个模式字符串和一组中间件函数作为参数。它会根据方法字符串和模式字符串生成一个键，并将其与通过 Chain 函数组合后的中间件函数作为值存储到 rules 字段中。
    - `Chain` 函数，用于将多个中间件函数组合在一起。它接受一组中间件函数作为参数，并返回一个新的中间件函数。当调用这个新的中间件函数时，它会依次调用所有传入的中间件函数，并将最终的处理器传递给最后一个中间件函数。
    - `ListenAndServe` 方法，用于启动服务器并监听指定的地址上的 TCP 连接。当接收到新的连接时，它会创建并初始化一个 server.Conn 实例，并调用 Serve 方法来处理这个连接。
    - `Serve` 方法，用于处理客户端连接。它会根据请求的方法和路径从 rules 字段中查找对应的处理器函数，并调用它来处理请求。如果没有找到匹配的处理器函数，则返回 404 Not Found 的响应。

<br/>

- `context` 模块用于解析和打印 HTTP 消息。它定义了一个 Context 结构体，它包含了三个字段：StartLine、Headers 和 Body，分别用于存储报文的起始行、头部字段和报文主体。

- `context` 模块提供了以下几个方法和函数：

    - `NewContext` 函数，用于从 Req 变量中创建一个 Context 实例，并返回它。该函数会从 Req 中解析出报文的各个部分，并存储到 Context 实例中。
    - `Print` 方法，用于打印 Context 实例的各个部分，方便调试。该方法使用 fmt 包提供的函数来格式化输出起始行、头部字段和报文主体。
    - `ReadFormData` 方法，用于从报文主体 Body 中读取 form-data，并返回一个 map 类型的结果和一个错误值。该方法会从头部字段 Headers 中获取内容类型（Content-Type），并解析出边界（boundary）的值。然后将报文主体 Body 转换为一个字节切片，并使用边界作为分隔符，将其分割成多个字节切片。每个字节切片代表一个表单数据，包含头部字段和值两个部分。然后调用 parseHeader 函数来解析头部字段，获取名称和值，并将它们存储到 map 中。
    - `parseHeader` 函数，用于解析头部字段，获取名称和值，并返回它们和一个错误值。该函数会将头部字段转换为字符串，并按照分号（;）分割成多个部分。然后遍历每个部分，查找以 name= 或 filename= 开头的部分，并获取它们的值。如果有文件名，就将文件名作为值的一部分，并去掉前后的回车换行符（CRLF）。

<br/>

- `server` 模块用于实现 HTTP 和 WebSocket 的服务。它定义了一个 Conn 结构体，它包含了五个字段：Conn、Message、Data、mu 和 ws。Conn 字段是一个 net.Conn 类型的值，表示服务器和客户端之间的 TCP 连接。Message 字段是一个 context.Context 类型的指针，表示解析后的 HTTP 消息。Data 字段是一个 map 类型的值，用于存储和检索与连接相关的数据。mu 字段是一个 sync.RWMutex 类型的值，用于对连接进行读写锁定。ws 字段是一个 bool 类型的值，表示是否升级为 WebSocket 连接。

- `server` 模块提供了以下几个方法和函数：

    - `Set` 方法，用于跨中间件设置值。它接受一个字符串类型的键和一个空接口类型的值作为参数。它会在 Data 字段中存储键值对，并使用 mu 字段对连接进行写锁定。
    - `Get` 方法，用于跨中间件获取值。它接受一个字符串类型的键作为参数，并返回一个空接口类型的值和一个布尔类型的值作为结果。它会从 Data 字段中检索键对应的值，并使用 mu 字段对连接进行读锁定。
    - `WriteResponse` 方法，用于将一个自定义的 HTTP 响应写入到 Conn 中。它接受四个参数：statusCode、statusText、body 和 headers。statusCode 是一个整数类型的值，表示响应的状态码。statusText 是一个字符串类型的值，表示响应的原因短语。body 是一个字节切片类型的值，表示响应的主体。headers 是一个可变参数，表示响应的头部字段。该方法会使用 mu 字段对连接进行写锁定，并返回一个错误值。
    - `detectContentType` 函数，用于根据 body 的内容自动检测 MIME 类型，并返回一个字符串类型的结果。该函数会根据 body 的第一个字节判断是否是 HTML 或 JSON 类型，否则使用二进制流类型。
    - `IsWebSocket` 方法，用于返回 Conn 是否已经升级为一个 WebSocket 连接。它会从 Data 字段中获取 “websocket” 键对应的值，并返回它。如果 Data 字段为空，就先创建一个空的 map。
    - `UpgradeToWebSocket` 方法，用于将一个 Conn 升级为一个 WebSocket 连接，通过进行一个握手。它会使用 mu 字段对连接进行写锁定，并返回一个错误值。该方法会根据请求消息中的头部字段判断是否符合 WebSocket 协议的要求，并生成响应消息中的 Sec-WebSocket-Accept 头。然后将响应消息写入到 Conn 中，并将 Data 字段中的 “websocket” 键设置为 true，表示已经升级为 WebSocket 连接。
    - `ReadWebSocketMessage` 方法，用于从一个 WebSocket 连接中读取一个消息，并返回它的操作码、有效载荷和错误值。该方法会使用 mu 字段对连接进行读锁定。该方法会创建一个 bufio.Reader 类型的缓冲读取器，并使用 readWebSocketFrame 函数来读取帧，直到遇到最后一个帧为止。如果操作码是关闭帧，则返回操作码、空有效载荷和 EOF 错误。如果操作码是 ping 帧，则发送一个 pong 帧给对方，并继续循环。如果操作码是 pong 帧，则忽略它，并继续循环。
    - `readWebSocketFrame` 函数，用于从一个 WebSocket 连接中读取一个帧，并返回它的 fin 位、操作码和有效载荷和错误值。该函数会从缓冲读取器中读取第一和第二个字节，并获取 fin 位、操作码、MASK 位和有效载荷长度的值。如果有效载荷长度为 126 或 127，则表示后面有扩展长度，需要再从读取器中读取两个或八个字节，并将它们合并为扩展长度的值。如果有效载荷长度超过限制，则返回错误值。然后判断 MASK 位是否为 true，如果是，则表示后面有四个字节是掩码，需要再从读取器中读取四个字节，并存储到数组中。然后创建一个切片 payload，用于存储有效载荷。从读取器中读取有效载荷长度的数据，并存储到切片中。如果出错，则返回错误值。最后判断 MASK 位是否为 true，如果是，则表示需要对有效载荷进行异或运算。遍历有效载荷的每个字节，并与掩码的对应字节进行异或运算。然后返回 fin 位、操作码、有效载荷和 nil 错误值。
    - `WriteWebSocketMessage` 方法，用于将一个消息写入到 WebSocket 连接中。它接受两个参数：opCode 和 payload。opCode 是一个整数类型的值，表示消息的操作码。payload 是一个字节切片类型的值，表示消息的有效载荷。它会使用 mu 字段对连接进行写锁定，并返回一个错误值。该方法会创建一个 bytes.Buffer 类型的缓冲区，用于存放 WebSocket 帧。然后设置帧的第一个字节，包含 fin 位和操作码。假设消息没有分片，所以 fin 位为 1。然后设置帧的第二个字节，包含 mask 位和有效载荷长度。不使用掩码，所以 mask 位为 0。根据有效载荷长度的大小，选择合适的编码方式，并将长度字段和扩展长度（如果有）写入到缓冲区中。然后写入有效载荷，不进行掩码操作。最后将缓冲区的内容写入到网络连接中，并返回错误值（如果有）。
    - `CloseWebSocket` 方法，用于关闭 WebSocket 连接。它会使用 mu 字段对连接进行写锁定，并返回一个错误值。该方法会判断 Conn 是否已经升级为一个 WebSocket 连接，如果不是，则返回错误。然后发送一个关闭帧给对方，并等待对方回复一个关闭帧。然后关闭底层的 net.Conn，并将 Data 字段中的 “websocket” 键设置为 false，表示已经关闭 WebSocket 连接。
    - `WebSocketHandleError` 函数，用于处理读取或写入 WebSocket 消息时发生的错误。它接受一个错误类型的参数 err，并没有返回值。该函数会判断错误是否是 EOF，如果是，则表示对方关闭了连接，并打印相应的信息。然后判断错误是否是一个网络错误，并且是否是超时错误，如果是，则表示连接超时，并打印相应的信息。其他情况下，表示发生了意外的错误，并打印相应的信息。最后调用 CloseWebSocket 方法关闭连接。

<br/>

上面的代码框架实现了 HTTP 和 WebSocket 的基本功能，但还有一些可以改进或扩展的地方：

- 对于 HTTP 的服务，可以增加更多的头部字段和主体类型来支持不同的内容格式和编码方式，例如 JSON、XML、multipart/form-data 等。
- 对于 HTTP 的服务，可以增加更多的状态码和原因短语来响应不同的情况，例如 200 OK、301 Moved Permanently、400 Bad Request、500 Internal Server Error 等。
- 对于 HTTP 的服务，可以增加更多的中间件函数来实现一些通用或特定的功能，例如日志记录、身份验证、压缩、缓存等。
- 对于 WebSocket 的服务，可以增加对分片消息的支持，即将一个大的消息分成多个帧发送，并在接收端重新组合。
- 对于 WebSocket 的服务，可以增加对控制帧的支持，即在发送或接收数据帧的同时，也可以发送或接收关闭帧、ping 帧或 pong 帧。
- 对于 WebSocket 的服务，可以增加对文本和二进制帧的不同处理方式，例如对文本帧进行编码或解码，对二进制帧进行压缩或解压等。
- 对于 WebSocket 的服务，可以增加更多的错误处理和异常处理，例如对连接断开、超时、格式错误等情况进行恰当的响应或重连。

对于该框架的用户，需要考虑以下几点：

- 用户需要了解 HTTP 和 WebSocket 协议的基本原理和规范，以便正确地构造请求和响应消息，并处理不同的操作码和有效载荷。
- 用户需要使用 router 模块提供的方法和函数来定义路由规则和处理器函数，并使用 context 模块提供的方法和函数来解析和打印 HTTP 消息。
- 用户需要使用 server 模块提供的方法和函数来设置和获取与连接相关的数据，并根据需要升级或关闭 WebSocket 连接，并读取或写入 WebSocket 消息。