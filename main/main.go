package main

import (
	"fmt"
	"github.com/lvkeliang/httpws/middleware"
	"github.com/lvkeliang/httpws/router"
	"github.com/lvkeliang/httpws/server"
	"io"
	"log"
)

// 下面是一个简单的使用示例，它展示了如何使用我之前提供的五个模块来构建一个简单的 Web 应用：

func main() {
	r := router.NewRouter()
	r.HandleFunc("GET", "/", indexHandler)
	r.HandleFunc("POST", "/hello", middleware.Chain(loggingMiddleware, nameMiddleware, printFormData, helloMiddleware))

	s := &server.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Starting server on :8080")
	s.ListenAndServe()
}

func indexHandler(c server.Conn) {
	io.WriteString(c.Conn, "HTTP/1.1 200 OK\r\n\r\nWelcome to my website!")
	// c.Message.Print()
}

func helloMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		name, ok := c.Get("name")
		if !ok {
			name = "World"
		}
		io.WriteString(c.Conn, fmt.Sprintf("HTTP/1.1 200 OK\r\n\r\nHello, %s!", name))
	}
}

func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		// log.Printf("Received request: %s\n", c.Req)
		// c.Message.Print()
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

// 这个示例代码定义了两个路由处理器：indexHandler 和 helloHandler。indexHandler 处理器用于处理根路径 / 的请求，它会向客户端发送一条欢迎消息。helloHandler 处理器用于处理 /hello 路径的请求，它会从上下文对象中获取 name 数据，并向客户端发送一条问候消息。

// 这个示例代码还定义了一个中间件函数：loggingMiddleware。这个中间件函数用于在处理请求之前打印一条日志消息，记录收到的请求数据。

// 最后，这个示例代码创建了一个 Server 实例，并使用 ListenAndServe 方法启动服务器。服务器监听 :8080 端口上的 TCP 连接，并使用路由器和中间件来处理客户端发送的请求。
