package server

import (
	"github.com/lvkeliang/httpws/message"
	"io"
	"log"
	"net"
	"sync"
)

// 您好！您的想法非常有趣。这是一个简单的服务器模块包的实现示例，它使用了 net 包来监听端口并接收客户端连接：

type Server struct {
	Addr    string
	Handler Handler
}

type Handler interface {
	Serve(c *Conn)
}

type Conn struct {
	Conn    net.Conn
	Message *message.Message
	Data    map[string]interface{}
	mu      sync.RWMutex
}

func (c *Conn) Set(key string, value interface{}) {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data[key] = value
}

func (c *Conn) Get(key string) (value interface{}, ok bool) {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.Data[key]
	return
}

func (s *Server) ListenAndServe() {
	c := new(Conn)

	listener, err := net.Listen("tcp", s.Addr)
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
			c.Message, err = message.NewMessage(req[:n])
			if err != nil {
				log.Println("create new message err: ", err)
				return
			}
			s.Handler.Serve(c)
			c.Conn.Close()
		}()
	}
}

// 这个服务器模块包定义了一个 Server 结构体，它包含了服务器监听的地址和一个处理器接口。ListenAndServe 方法使用 net.Listen 函数监听指定的地址上的 TCP 连接，当接收到新的连接时，它会调用处理器的 Serve 方法来处理这个连接。
