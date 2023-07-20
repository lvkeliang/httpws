// 这是一个简单的路由模块包的实现示例，它定义了一个 Router 结构体，用于存储路由规则和处理器：

package router

import (
	"github.com/lvkeliang/httpws/server"
	"io"
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

func (r *Router) HandleFunc(pattern string, handler HandlerFunc) {
	r.rules[pattern] = handler
}

func (r *Router) Serve(c *server.Conn) {
	reqStr := string(c.Req)
	path := strings.Split(reqStr, " ")[1]
	handler, ok := r.rules[path]
	if !ok {
		io.WriteString(c.Conn, "HTTP/1.1 404 Not Found\r\n\r\nNot Found")
		return
	}
	handler(*c)
}

// 这个路由模块定义了一个 Router 结构体，它包含了一个 rules 字段，用于存储路由规则和处理器。HandleFunc 方法用于添加新的路由规则，它接受一个模式字符串和一个处理器函数作为参数。Serve 方法用于处理客户端连接，它会根据请求的 URL 路径查找对应的处理器，并调用它来处理请求。
