package server

import (
	"github.com/lvkeliang/httpws/context"
	"io"
	"log"
	"net"
)

// 您好！您的想法非常有趣。这是一个简单的服务器模块包的实现示例，它使用了 net 包来监听端口并接收客户端连接：

type Server struct {
	Addr    string
	Handler Handler
}

type Handler interface {
	Serve(conn net.Conn, req []byte)
}

type connServe struct {
	listener net.Listener
	conn     net.Conn
	context  context.Context
}

func (s *Server) ListenAndServe() {
	connectServe := new(connServe)

	var err error
	connectServe.listener, err = net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer connectServe.listener.Close()

	for {
		connectServe.conn, err = connectServe.listener.Accept()
		if err != nil {
			log.Println("listener err: ", err)
			continue
		}
		go func() {
			req := make([]byte, 1024)
			n, err := connectServe.conn.Read(req)
			if err != nil {
				if err != io.EOF {
					log.Println("conn read err: ", err)
				}
				return
			}
			s.Handler.Serve(connectServe.conn, req[:n])
			connectServe.conn.Close()
		}()
	}
}

// 这个服务器模块包定义了一个 Server 结构体，它包含了服务器监听的地址和一个处理器接口。ListenAndServe 方法使用 net.Listen 函数监听指定的地址上的 TCP 连接，当接收到新的连接时，它会调用处理器的 Serve 方法来处理这个连接。
