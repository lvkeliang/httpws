// Package middleware 这个中间件模块定义了一个 Middleware 类型，它是一个函数类型，接受一个 HandlerFunc 类型的参数，并返回一个新的 HandlerFunc 类型的值。
// Chain 函数用于将多个中间件函数组合在一起，它接受一组中间件函数作为参数，并返回一个新的中间件函数。当调用这个新的中间件函数时，它会依次调用所有传入的中间件函数，并将最终的处理器传递给最后一个中间件函数。
package middleware

import (
	"github.com/lvkeliang/httpws/router"
	"github.com/lvkeliang/httpws/server"
)

type Middleware func(router.HandlerFunc) router.HandlerFunc

func Chain(middlewares ...Middleware) router.HandlerFunc {

	chain := func(final router.HandlerFunc) router.HandlerFunc { // 修改这一行
		return func(c server.Conn) {
			last := final
			for i := len(middlewares) - 1; i >= 0; i-- {
				last = middlewares[i](last)
			}

			if last != nil { // 添加这一行，检查 last 是否为 nil
				last(c)
			}
		}
	}

	return chain(nil)
}
