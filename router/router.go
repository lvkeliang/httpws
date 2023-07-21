// Package router 这个路由模块定义了一个 Router 结构体，它包含了一个 rules 字段，用于存储路由规则和处理器。
// HandleFunc 方法用于添加新的路由规则，它接受一个模式字符串和一个处理器函数作为参数。
// Serve 方法用于处理客户端连接，它会根据请求的 URL 路径查找对应的处理器，并调用它来处理请求。
package router

import (
	"github.com/lvkeliang/httpws/server"
	"log"
	"strings"
)

type HandlerFunc func(c server.Conn)

//// 实现Handler接口
//func (f HandlerFunc) Serve(conn net.Conn, req []byte) {
//	f(conn, req)
//}

type Router struct {
	rules map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		rules: make(map[string]HandlerFunc),
	}
}

func (r *Router) HandleFunc(method string, pattern string, handler HandlerFunc) {
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
