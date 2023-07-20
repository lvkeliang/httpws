// 这是一个简单的基于 net 包的上下文模块包的实现示例，它定义了一个 Context 结构体，用于在处理请求的过程中传递数据：

package context

import (
	"net"
	"sync"
)

type Context struct {
	Conn net.Conn
	Req  []byte
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewContext(conn net.Conn, req []byte) *Context {
	return &Context{
		Conn: conn,
		Req:  req,
		data: make(map[string]interface{}),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Context) Get(key string) (value interface{}, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.data[key]
	return
}

// 这个上下文模块包定义了一个 Context 结构体，它包含了客户端连接和请求数据两个字段，以及一个用于存储自定义数据的 data 字段。Set 方法用于设置自定义数据，它接受一个键和一个值作为参数。Get 方法用于获取自定义数据，它接受一个键作为参数，并返回对应的值和一个布尔值，表示是否找到了对应的键。
