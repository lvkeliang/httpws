package main

import (
	"fmt"
	"github.com/lvkeliang/httpws/middleware"
	"github.com/lvkeliang/httpws/router"
	"github.com/lvkeliang/httpws/server"
	"log"
)

// 下面是一个简单的使用示例，它展示了如何使用该框架来构建一个简单的 Web 应用：
func main() {
	// 创建一个路由
	r := router.NewRouter()
	r.HandleFunc("GET", "/", indexHandler)
	r.HandleFunc("POST", "/hello", middleware.Chain(loggingMiddleware, nameMiddleware, printFormData, helloMiddleware))
	r.HandleFunc("GET", "/ws", middleware.Chain(handleWebSocket))

	// 创建了一个 Server 实例，并使用 ListenAndServe 方法启动服务器。服务器监听 :8080 端口上的 TCP 连接，并使用路由器和中间件来处理客户端发送的请求。
	s := &server.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Starting server on :8080")
	s.ListenAndServe()
}

func indexHandler(c server.Conn) {
	c.Message.Print()
	c.WriteResponse(200, "OK", []byte("Welcome to my website!"))
}

// 用于在处理请求之前打印一条日志消息，记录收到的请求数据。
func loggingMiddleware(next router.HandlerFunc) router.HandlerFunc {
	return func(c server.Conn) {
		c.Message.Print()
		// 不是最后一个中间件，需要调用next(c)
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
	}
	// 是最后一个中间件，不要调用next(c)
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
			fmt.Printf("Received message: opCode = %d, payload = %s\n", opCode, string(payload))
			// 将收到的消息原样写回到WebSocket连接中，如果出错，处理错误并退出循环
			if err := c.WriteWebSocketMessage(opCode, payload); err != nil {
				c.WebSocketHandleError(err)
				break
			}
		}
	}
}
