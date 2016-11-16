package rpcClient

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/Jordanzuo/goutil/intAndBytesUtil"
	"github.com/Jordanzuo/goutil/logUtil"
)

const (
	// 包头的长度
	con_HEADER_LENGTH = 4

	// 定义请求、响应数据的前缀的长度
	con_ID_LENGTH = 4
)

var (
	// 字节的大小端顺序
	byterOrder = binary.LittleEndian
)

// 定义客户端对象，以实现对客户端连接的封装
type client struct {
	// 客户端连接对象
	conn net.Conn

	// 接收到的消息内容
	content []byte
}

// 追加内容
// content：新的内容
// 返回值：无
func (clientObj *client) appendContent(content []byte) {
	clientObj.content = append(clientObj.content, content...)
}

// 获取有效的消息
// 返回值：
// 消息对应客户端的唯一标识
// 消息内容
// 是否含有有效数据
func (clientObj *client) getValidMessage() (int32, []byte, bool) {
	// 判断是否包含头部信息
	if len(clientObj.content) < con_HEADER_LENGTH {
		return 0, nil, false
	}

	// 获取头部信息
	header := clientObj.content[:con_HEADER_LENGTH]

	// 将头部数据转换为内部的长度
	contentLength := intAndBytesUtil.BytesToInt32(header, byterOrder)

	// 判断长度是否满足
	if len(clientObj.content) < con_HEADER_LENGTH+int(contentLength) {
		return 0, nil, false
	}

	// 提取消息内容
	content := clientObj.content[con_HEADER_LENGTH : con_HEADER_LENGTH+contentLength]

	// 将对应的数据截断，以得到新的数据
	clientObj.content = clientObj.content[con_HEADER_LENGTH+contentLength:]

	// 判断是否为心跳包，如果是心跳包，则不解析，直接返回
	if contentLength == 0 || len(content) == 0 {
		return 0, nil, false
	}

	// 判断内容的长度是否足够
	if len(content) < con_ID_LENGTH {
		logUtil.Log(fmt.Sprintf("内容数据不正确；con_ID_LENGTH=%d,len(content)=%d", con_ID_LENGTH, len(content)), logUtil.Warn, true)
		return 0, nil, false
	}

	// 截取内容的前4位
	idBytes, content := content[:con_ID_LENGTH], content[con_ID_LENGTH:]

	// 提取id
	id := intAndBytesUtil.BytesToInt32(idBytes, byterOrder)

	return id, content, true
}

// 发送字节数组消息
// id：需要添加到b前发送的数据
// message：待发送的字节数组
func (clientObj *client) sendByteMessage(id int32, message []byte) {
	idBytes := intAndBytesUtil.Int32ToBytes(id, byterOrder)

	// 将idByte和b合并
	message = append(idBytes, message...)

	// 获得数组的长度
	contentLength := len(message)

	// 将长度转化为字节数组
	header := intAndBytesUtil.Int32ToBytes(int32(contentLength), byterOrder)

	// 将头部与内容组合在一起
	message = append(header, message...)

	// 发送消息
	clientObj.conn.Write(message)
}

// 新建客户端对象
// conn：连接对象
// 返回值：客户端对象的指针
func newClient(_conn net.Conn) *client {
	return &client{
		conn:    _conn,
		content: make([]byte, 0, 1024),
	}
}
