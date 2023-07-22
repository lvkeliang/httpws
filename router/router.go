// Package router 这个路由模块定义了一个 Router 结构体，它包含了一个 rules 字段，用于存储路由规则和处理器。
package router

import (
	"github.com/lvkeliang/httpws/context"
	"github.com/lvkeliang/httpws/server"
	"io"
	"log"
	"net"
	"strings"
)

type HandlerFunc func(c server.Conn)

type Router struct {
	rules map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		rules: make(map[string]HandlerFunc),
	}
}

type Middleware func(HandlerFunc) HandlerFunc

// HandleFunc 方法用于添加新的路由规则，它接受一个模式字符串和一个处理器函数作为参数。
func (r *Router) HandleFunc(method string, pattern string, middlewares ...Middleware) {
	handler := Chain(middlewares)
	switch method {
	case "GET":
		r.rules[method+" "+pattern] = handler
	case "POST":
		r.rules[method+" "+pattern] = handler
	case "PUT":
		r.rules[method+" "+pattern] = handler
	case "PATCH":
		r.rules[method+" "+pattern] = handler
	case "DELETE":
		r.rules[method+" "+pattern] = handler
	case "HEAD":
		r.rules[method+" "+pattern] = handler
	case "OPTIONS":
		r.rules[method+" "+pattern] = handler
	default:
		log.Printf("method err: unsolved method \"%v\"\n", method)
	}
}

// Chain 函数用于将多个中间件函数组合在一起，它接受一组中间件函数作为参数，并返回一个新的中间件函数。
// 当调用这个新的中间件函数时，它会依次调用所有传入的中间件函数，并将最终的处理器传递给最后一个中间件函数。
func Chain(middlewares []Middleware) HandlerFunc {

	// 定义一个 chain 函数，它接受一个最终处理器作为参数，并返回一个新的处理器。\
	return func(c server.Conn) {
		// 定义最后的处理器是什么也不做
		var last = func(c server.Conn) {}

		// 逆序遍历 middlewares 切片。
		for i := len(middlewares) - 1; i >= 0; i-- {
			// 通过将最后的处理器应用于当前的中间件函数来更新它。
			last = middlewares[i](last)
		}

		// 使用服务器连接调用最后的处理器。
		last(c)
	}
}

// ListenAndServe 方法使用 net.Listen 函数监听指定的地址上的 TCP 连接，当接收到新的连接时，它会调用处理器的 Serve 方法来处理这个连接。
func (r *Router) ListenAndServe(addr string) {
	c := new(server.Conn)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		c.Conn, err = listener.Accept()
		if err != nil {
			log.Println("listener err: ", err)
			continue
		}
		go func() {
			req := make([]byte, 1024)
			n, err := c.Conn.Read(req)
			if err != nil {
				if err != io.EOF {
					log.Println("conn read err: ", err)
				}
				return
			}
			c.Message, err = context.NewContext(req[:n])
			if err != nil {
				log.Println("create new context err: ", err)
				return
			}
			r.Serve(c)
			c.Conn.Close()
		}()
	}
}

// Serve 方法用于处理客户端连接，它会根据请求的 URL 路径查找对应的处理器，并调用它来处理请求。
func (r *Router) Serve(c *server.Conn) {

	// 获取请求方法和路径，并按照请求的方法和路径调用中间件
	lsatInd := strings.LastIndex(c.Message.StartLine, " ")
	handler, ok := r.rules[c.Message.StartLine[:lsatInd]]
	if !ok {
		c.WriteResponse(404, "404 Not Found", []byte("Not Found"))
		return
	}
	handler(*c)
}
