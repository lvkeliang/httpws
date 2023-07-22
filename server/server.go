// Package server 这个服务器模块用于实现HTTP和WebSocket服务，包定义了一个 Server 结构体，它包含了服务器监听的地址和一个处理器接口。
package server

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lvkeliang/httpws/context"
	"io"
	"log"
	"math"
	"net"
	"strings"
	"sync"
)

type Conn struct {
	Conn    net.Conn
	Message *context.Context
	Data    map[string]interface{}
	mu      sync.RWMutex
}

// Set 用于跨中间件设置值
func (c *Conn) Set(key string, value interface{}) {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data[key] = value
}

// Get 用于跨中间件设置值
func (c *Conn) Get(key string) (value interface{}, ok bool) {
	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok = c.Data[key]
	return
}

// WriteResponse 将一个自定义的http响应写入到Conn中
func (c *Conn) WriteResponse(statusCode int, statusText string, body []byte, headers ...map[string]string) error {
	// 对Conn加写锁
	c.mu.Lock()
	defer c.mu.Unlock()

	// 创建一个缓冲区来写入响应
	var buf bytes.Buffer

	// 写入状态行
	fmt.Fprintf(&buf, "HTTP/1.1 %d %s\r\n", statusCode, statusText)

	// 根据body的内容自动检测MIME类型
	contentType := detectContentType(body)

	// 写入内容类型头
	fmt.Fprintf(&buf, "Content-Type: %s\r\n", contentType)

	// 写入内容长度头
	fmt.Fprintf(&buf, "Content-Length: %d\r\n", len(body))

	// 写入用户自定义的其他头部，如果有的话
	for _, header := range headers {
		for key, value := range header {
			// fmt.Printf("headers: %s: %s\r\n", key, value)
			fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
		}
	}

	// 写入一个空行来分隔头部和主体
	fmt.Fprint(&buf, "\r\n")

	// 写入主体
	buf.Write(body)

	// 将缓冲区的内容写入到Conn中
	_, err := c.Conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// detectContentType 根据body的内容自动检测MIME类型
func detectContentType(body []byte) string {
	// 如果body为空，返回默认的文本类型
	if len(body) == 0 {
		return "text/plain; charset=utf-8"
	}

	// 如果body以"<"开头，假设它是HTML类型
	if body[0] == '<' {
		return "text/html; charset=utf-8"
	}

	// 如果body以"{"或"["开头，假设它是JSON类型
	if body[0] == '{' || body[0] == '[' {
		return "application/json"
	}

	// 其他情况，使用二进制流类型
	return "application/octet-stream"
}

const (
	// WebSocketMagicString 是用于WebSocket握手的一个魔术字符串
	WebSocketMagicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	// WebSocketVersion 是WebSocket协议的版本
	WebSocketVersion = "13"

	// WebSocketFrameFinBit 是用于表示FIN位的位掩码，在WebSocket帧的第一个字节中
	WebSocketFrameFinBit = 1 << 7

	// WebSocketFrameOpCodeMask 是用于表示操作码的位掩码，在WebSocket帧的第一个字节中
	WebSocketFrameOpCodeMask = 0x0F

	// WebSocketFrameOpCodeText 是用于表示文本帧的操作码
	WebSocketFrameOpCodeText = 0x01

	// WebSocketFrameOpCodeBinary 是用于表示二进制帧的操作码
	WebSocketFrameOpCodeBinary = 0x02

	// WebSocketFrameOpCodeClose 是用于表示关闭帧的操作码
	WebSocketFrameOpCodeClose = 0x08

	// WebSocketFrameOpCodePing 是用于表示ping帧的操作码
	WebSocketFrameOpCodePing = 0x09

	// WebSocketFrameOpCodePong 是用于表示pong帧的操作码
	WebSocketFrameOpCodePong = 0x0A

	// WebSocketFrameMaskBit 是用于表示MASK位的位掩码，在WebSocket帧的第二个字节中
	WebSocketFrameMaskBit = 1 << 7

	// WebSocketFramePayloadLenMask 是用于表示有效载荷长度的位掩码，在WebSocket帧的第二个字节中
	WebSocketFramePayloadLenMask = 0x7F

	// WebSocketMaxPayloadLen 是WebSocket帧的最大有效载荷长度
	WebSocketMaxPayloadLen = 1<<63 - 1
)

var (
	errInvalidHandshake    = errors.New("invalid handshake")
	errUnsupportedProtocol = errors.New("unsupported protocol")
	errInvalidFrame        = errors.New("invalid frame")
)

// IsWebSocket 返回Conn是否已经升级为一个WebSocket连接，是则返回true，否则返回false
func (c *Conn) IsWebSocket() bool {
	// c.mu.RLock() // 对Conn加读锁
	// defer c.mu.RUnlock()

	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	return c.Data["websocket"] == true // 返回c.Data["websocket"]的值
}

// UpgradeToWebSocket 将一个Conn升级为一个WebSocket连接，通过进行一个握手
func (c *Conn) UpgradeToWebSocket() error {
	c.mu.Lock() // 对Conn加写锁
	defer c.mu.Unlock()

	if c.Message == nil { // 如果没有收到消息，返回错误
		log.Println("Context == nil")
		return errInvalidHandshake
	}

	if !strings.HasPrefix(c.Message.StartLine, "GET") || !strings.HasSuffix(c.Message.StartLine, "HTTP/1.1") { // 如果请求行不是GET / HTTP/1.1，返回错误
		log.Printf("Context.StartLine != \"GET / HTTP/1.1\"\nreceved: %v\n", c.Message.StartLine)
		return errInvalidHandshake
	}

	if c.Message.Headers["Upgrade"] != "websocket" { // 如果Upgrade头不是websocket，返回错误
		return errInvalidHandshake
	}

	if c.Message.Headers["Connection"] != "Upgrade" { // 如果Connection头不是Upgrade，返回错误
		return errInvalidHandshake
	}

	if c.Message.Headers["Sec-WebSocket-Version"] != WebSocketVersion { // 如果Sec-WebSocket-Version头不是13，返回错误
		return errUnsupportedProtocol
	}

	key := c.Message.Headers["Sec-WebSocket-Key"] // 获取Sec-WebSocket-Key头的值
	if key == "" {                                // 如果没有这个头，返回错误
		return errInvalidHandshake
	}

	hash := sha1.Sum([]byte(key + WebSocketMagicString))      // 对key和魔术字符串进行SHA1哈希
	responseKey := base64.StdEncoding.EncodeToString(hash[:]) // 对哈希结果进行Base64编码

	response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", responseKey) // 构造响应消息

	if _, err := c.Conn.Write([]byte(response)); err != nil { // 将响应消息写入到Conn中，如果出错，返回错误
		return err
	}

	if c.Data == nil {
		c.Data = make(map[string]interface{})
	}
	c.Data["websocket"] = true // 将c.Data["websocket"]设置为true，表示已经升级为WebSocket连接

	return nil // 返回nil表示成功
}

// ReadWebSocketMessage 从一个WebSocket连接中读取一个消息，并返回它的操作码和有效载荷
func (c *Conn) ReadWebSocketMessage() (int, []byte, error) {
	c.mu.RLock() // 对Conn加读锁
	defer c.mu.RUnlock()

	if !c.IsWebSocket() { // 如果不是一个WebSocket连接，返回错误
		return 0, nil, errors.New("not a websocket connection")
	}

	reader := bufio.NewReader(c.Conn) // 创建一个缓冲读取器

	var opCode int     // 声明一个变量用于存储操作码
	var payload []byte // 声明一个切片用于存储有效载荷

	for {
		fin, op, data, err := readWebSocketFrame(reader) // 从读取器中读取一个帧，并获取它的fin位、操作码、有效载荷和错误
		if err != nil {                                  // 如果出错，返回错误
			if err != io.EOF {
				fmt.Println(err)
				return 0, nil, err
			}
		}

		if op == WebSocketFrameOpCodeClose { // 如果操作码是关闭帧，返回操作码、空有效载荷和EOF错误
			return op, nil, io.EOF
		}

		if op == WebSocketFrameOpCodePing { // 如果操作码是ping帧，发送一个pong帧给对方，并继续循环
			if err := c.WriteWebSocketMessage(WebSocketFrameOpCodePong, nil); err != nil {
				return 0, nil, err
			}
			continue
		}

		if op == WebSocketFrameOpCodePong { // 如果操作码是pong帧，忽略它，并继续循环
			continue
		}

		if opCode == 0 { // 如果操作码还没有被赋值，将它设置为当前帧的操作码
			opCode = op
		}

		payload = append(payload, data...) // 将当前帧的有效载荷追加到总的有效载荷中

		if fin { // 如果fin位为true，表示这是最后一个帧，跳出循环
			break
		}
	}

	return opCode, payload, nil // 返回操作码、有效载荷和nil错误
}

// readWebSocketFrame 从一个WebSocket连接中读取一个帧，并返回它的fin位、操作码和有效载荷
func readWebSocketFrame(reader *bufio.Reader) (bool, int, []byte, error) {
	b1, err := reader.ReadByte() // 读取第一个字节
	if err != nil {              // 如果出错，返回错误
		return false, 0, nil, err
	}

	fin := b1&WebSocketFrameFinBit != 0          // 获取fin位的值
	opCode := int(b1 & WebSocketFrameOpCodeMask) // 获取操作码的值

	b2, err := reader.ReadByte() // 读取第二个字节
	if err != nil {              // 如果出错，返回错误
		return false, 0, nil, err
	}

	masked := b2&WebSocketFrameMaskBit != 0                // 获取MASK位的值
	payloadLen := int64(b2 & WebSocketFramePayloadLenMask) // 获取有效载荷长度的值

	if payloadLen == 126 { // 如果有效载荷长度为126，表示后面两个字节是扩展长度
		b1, err := reader.ReadByte() // 读取第三个字节
		if err != nil {              // 如果出错，返回错误
			return false, 0, nil, err
		}
		b2, err := reader.ReadByte() // 读取第四个字节
		if err != nil {              // 如果出错，返回错误
			return false, 0, nil, err
		}
		payloadLen = int64(b1)<<8 | int64(b2) // 将两个字节合并为扩展长度的值
	} else if payloadLen == 127 { // 如果有效载荷长度为127，表示后面八个字节是扩展长度
		var b [8]byte
		if _, err := io.ReadFull(reader, b[:]); err != nil { // 读取后面八个字节到数组中，如果出错，返回错误
			return false, 0, nil, err
		}
		payloadLen = int64(b[0])<<56 | int64(b[1])<<48 | int64(b[2])<<40 | int64(b[3])<<32 |
			int64(b[4])<<24 | int64(b[5])<<16 | int64(b[6])<<8 | int64(b[7]) // 将八个字节合并为扩展长度的值
	}

	if payloadLen > WebSocketMaxPayloadLen { // 如果有效载荷长度超过限制，返回错误
		return false, 0, nil, errors.New("payload length exceeds limit")
	}

	var mask [4]byte
	if masked { // 如果MASK位为true，表示后面四个字节是掩码
		if _, err := io.ReadFull(reader, mask[:]); err != nil { // 读取后面四个字节到数组中，如果出错，返回错误
			return false, 0, nil, err
		}
	}

	payload := make([]byte, payloadLen)                     // 创建一个切片用于存储有效载荷
	if _, err := io.ReadFull(reader, payload); err != nil { // 读取有效载荷到切片中，如果出错，返回错误
		return false, 0, nil, err
	}

	if masked { // 如果MASK位为true，表示需要对有效载荷进行异或运算
		for i := range payload { // 遍历有效载荷的每个字节
			payload[i] ^= mask[i%4] // 与掩码的对应字节进行异或运算
		}
	}

	return fin, opCode, payload, nil // 返回fin位、操作码、有效载荷和nil错误
}

// WriteWebSocketMessage 将一个消息写入到连接中。
func (c *Conn) WriteWebSocketMessage(opCode int, payload []byte) error {
	// 锁定连接，防止并发写入。
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsWebSocket() { // 如果不是一个WebSocket连接，返回错误
		return errors.New("not a websocket connection")
	}

	// 创建一个缓冲区，用于存放websocket帧。
	var buf bytes.Buffer

	// 设置帧的第一个字节，包含fin位和操作码。
	fin := 1 // 假设消息没有分片。
	buf.WriteByte(byte(fin)<<7 | byte(opCode))

	// 设置帧的第二个字节，包含mask位和负载长度。
	mask := 0                          // 不使用掩码。
	payloadLen := uint64(len(payload)) // 获取负载长度，并转换为uint64类型。

	if payloadLen < 126 {
		// 使用7位来编码长度。
		buf.WriteByte(byte(mask)<<7 | byte(payloadLen))
	} else if payloadLen <= math.MaxUint16 {
		// 使用16位来编码长度，并将长度字段设为126。
		buf.WriteByte(byte(mask)<<7 | 126)
		// 以网络字节序（大端）写入长度，使用uint16类型。
		binary.Write(&buf, binary.BigEndian, uint16(payloadLen))
	} else if payloadLen <= math.MaxUint32 {
		// 使用32位来编码长度，并将长度字段设为127。
		buf.WriteByte(byte(mask)<<7 | 127)
		// 以网络字节序（大端）写入长度，使用uint32类型。
		binary.Write(&buf, binary.BigEndian, uint32(payloadLen))
	} else {
		// 使用64位来编码长度，并将长度字段设为127。
		buf.WriteByte(byte(mask)<<7 | 127)
		// 以网络字节序（大端）写入长度，使用uint64类型。
		binary.Write(&buf, binary.BigEndian, payloadLen)
	}

	// 写入负载，不进行掩码操作。
	buf.Write(payload)

	// 将缓冲区写入到网络连接中。
	if _, err := c.Conn.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

// CloseWebSocket 关闭WebSocket连接
func (c *Conn) CloseWebSocket() error {
	c.mu.Lock() // 对Conn加写锁
	defer c.mu.Unlock()

	if !c.IsWebSocket() { // 如果不是一个WebSocket连接，返回错误
		return errors.New("not a websocket connection")
	}

	// Send a close frame to the peer 发送一个关闭帧给对方
	if err := c.WriteWebSocketMessage(WebSocketFrameOpCodeClose, nil); err != nil { // 如果出错，返回错误
		return err
	}

	// Wait for a close frame from the peer 等待对方回复一个关闭帧
	for {
		opCode, _, err := c.ReadWebSocketMessage() // 读取一个消息，并获取它的操作码和错误
		if err != nil {                            // 如果出错，返回错误
			return err
		}
		if opCode == WebSocketFrameOpCodeClose { // 如果操作码是关闭帧，跳出循环
			break
		}
	}

	// Close the underlying net.Conn 关闭底层的net.Conn
	if err := c.Conn.Close(); err != nil { // 如果出错，返回错误
		return err
	}

	c.Data["websocket"] = false // 将c.Data["websocket"]设置为false，表示已经关闭WebSocket连接

	return nil // 返回nil表示成功
}

// WebSocketHandleError 处理读取或写入WebSocket消息时发生的错误
func (c *Conn) WebSocketHandleError(err error) {
	if err == io.EOF { // 如果错误是EOF，表示对方关闭了连接
		fmt.Println("connection closed by peer")
	} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() { // 如果错误是一个网络错误，并且是超时错误，表示连接超时
		fmt.Println("connection timed out")
	} else { // 其他情况，表示发生了意外的错误
		fmt.Println("unexpected error:", err)
	}
	c.CloseWebSocket() // 调用Close方法关闭连接
}
