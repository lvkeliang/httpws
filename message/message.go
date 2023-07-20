// Package message 这是一个简单的读取报文数据的包，它定义了一个 Message 结构体，用于存储报文的各个部分：
package message

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Message struct {
	StartLine string            // 起始行
	Headers   map[string]string // 头部字段
	Body      []byte            // 报文主体
}

// NewMessage 函数用于从 Req 变量中创建一个 Message 实例，并返回它：
func NewMessage(Req []byte) (*Message, error) {
	m := &Message{}                            // 创建一个空的 Message 实例
	r := bufio.NewReader(bytes.NewReader(Req)) // 创建一个 Reader 对象，用于从 Req 变量中读取数据

	// 读取起始行
	startLine, err := r.ReadBytes('\n') // 读取直到遇到换行符（\n）为止
	if err != nil {
		return nil, err // 如果读取失败，返回错误
	}
	m.StartLine = string(startLine[:len(startLine)-2]) // 将起始行转换为字符串，并去掉最后的回车换行符（CRLF）

	// 读取头部字段
	m.Headers = make(map[string]string) // 创建一个空的 map，用于存储头部字段
	for {
		line, err := r.ReadBytes('\n') // 读取直到遇到换行符（\n）为止
		if err != nil {
			return nil, err // 如果读取失败，返回错误
		}
		if len(line) == 2 { // 如果只有两个字节，说明是空行，表示头部字段结束
			break // 跳出循环
		}
		parts := bytes.SplitN(line, []byte{':'}, 2) // 将每一行按照冒号（:）分割成两个部分
		if len(parts) != 2 {                        // 如果不是两个部分，说明格式错误
			return nil, errors.New("invalid header format") // 返回错误
		}
		name := string(parts[0])                                     // 第一个部分是头部字段的名称
		value := string(bytes.TrimSpace(parts[1][:len(parts[1])-2])) // 第二个部分是头部字段的值，需要去掉前后的空白字符和最后的回车换行符（CRLF）
		m.Headers[name] = value                                      // 将头部字段的名称和值存储在 map 中
	}

	// 读取报文主体
	contentLength, ok := m.Headers["Content-Length"] // 从头部字段中获取内容长度（Content-Length）
	if !ok {                                         // 如果没有内容长度，说明没有报文主体
		return m, nil // 返回 Message 实例
	}
	length, err := strconv.Atoi(contentLength) // 将内容长度转换为整数
	if err != nil {
		return nil, err // 如果转换失败，返回错误
	}
	m.Body = make([]byte, length)   // 创建一个指定长度的字节切片，用于存储报文主体
	_, err = io.ReadFull(r, m.Body) // 从 Reader 对象中读取指定长度的数据，存储到报文主体中
	if err != nil {
		return nil, err // 如果读取失败，返回错误
	}

	return m, nil // 返回 Message 实例
}

// Print 函数用于打印 Message 实例的各个部分，方便调试：
func (m *Message) Print() {
	fmt.Println("StartLine:", m.StartLine) // 打印起始行
	fmt.Println("Headers:")                // 打印头部字段
	for name, value := range m.Headers {
		fmt.Printf("%s: %s\n", name, value)
	}
	fmt.Println("Body:", string(m.Body)) // 打印报文主体
}
