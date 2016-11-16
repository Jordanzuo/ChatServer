package rpcServer

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Jordanzuo/ChatServerModel/src/serverResponseObject"
	"github.com/Jordanzuo/goutil/fileUtil"
	"github.com/Jordanzuo/goutil/intAndBytesUtil"
	"github.com/Jordanzuo/goutil/logUtil"
	"github.com/Jordanzuo/goutil/timeUtil"
)

const (
	// 包头的长度
	con_HEADER_LENGTH = 4
)

var (
	// 全局客户端的id，从1开始进行自增
	globalClientId int32 = 0

	// 字节的大小端顺序
	byterOrder = binary.LittleEndian
)

// 定义客户端对象，以实现对客户端连接的封装
type Client struct {
	// 唯一标识
	id int32

	// 客户端连接对象
	conn net.Conn

	//连接状态 (1:连接中, 2:最后一条消息，3,断开)
	connStatus ConnStatus

	// 接收到的消息内容
	receiveData []byte

	// 待发送的数据
	sendData []*serverResponseObject.ResponseObject

	// 低优先级的待发送的数据
	sendData_LowPriority []*serverResponseObject.ResponseObject

	// 锁对象（用于控制对sendDatap、sendData_LowPriority的并发访问；receiveData不需要，因为是同步访问）
	mutex sync.Mutex

	// 玩家Id
	playerId string

	// 上次活跃时间
	activeTime time.Time
}

// 获取唯一标识
func (c *Client) GetId() int32 {
	return c.id
}

// 获取玩家Id
// 返回值：
// 玩家Id
func (c *Client) GetPlayerId() string {
	return c.playerId
}

// 获取远程地址（IP_Port）
func (clientObj *Client) getRemoteAddr() string {
	items := strings.Split(clientObj.conn.RemoteAddr().String(), ":")

	return fmt.Sprintf("%s_%s", items[0], items[1])
}

// 获取远程地址（IP）
func (clientObj *Client) getRemoteShortAddr() string {
	items := strings.Split(clientObj.conn.RemoteAddr().String(), ":")

	return items[0]
}

// 获取有效的消息
// 返回值：
// 消息内容
// 是否含有有效数据
func (c *Client) getReceiveData() ([]byte, bool) {
	// 判断是否包含头部信息
	if len(c.receiveData) < con_HEADER_LENGTH {
		return nil, false
	}

	// 获取头部信息
	header := c.receiveData[:con_HEADER_LENGTH]

	// 将头部数据转换为内部的长度
	contentLength := intAndBytesUtil.BytesToInt32(header, byterOrder)

	// 判断长度是否满足
	if len(c.receiveData) < con_HEADER_LENGTH+int(contentLength) {
		return nil, false
	}

	// 提取消息内容
	content := c.receiveData[con_HEADER_LENGTH : con_HEADER_LENGTH+contentLength]

	// 将对应的数据截断，以得到新的数据
	c.receiveData = c.receiveData[con_HEADER_LENGTH+contentLength:]

	return content, true
}

// 获取待发送的数据
// 返回值：
// 待发送数据项
// 是否含有有效数据
func (clientObj *Client) getSendData() (responseObj *serverResponseObject.ResponseObject, exists bool) {
	clientObj.mutex.Lock()
	defer clientObj.mutex.Unlock()

	// 如果没有数据则直接返回
	if len(clientObj.sendData) == 0 {
		return
	}

	// 取出第一条数据,并为返回值赋值
	responseObj = clientObj.sendData[0]
	exists = true

	// 删除已经取出的数据
	clientObj.sendData = clientObj.sendData[1:]

	return
}

// 获取低优先级待发送的数据
// 返回值：
// 待发送数据项
// 是否含有有效数据
func (clientObj *Client) getSendData_LowPriority() (sendDataItemObj *serverResponseObject.ResponseObject, exists bool) {
	clientObj.mutex.Lock()
	defer clientObj.mutex.Unlock()

	// 如果没有数据则直接返回
	if len(clientObj.sendData_LowPriority) == 0 {
		return
	}

	// 取出第一条数据,并为返回值赋值
	sendDataItemObj = clientObj.sendData_LowPriority[0]
	exists = true

	// 删除已经取出的数据
	clientObj.sendData_LowPriority = clientObj.sendData_LowPriority[1:]

	return
}

// 获取连接状态
func (clientObj *Client) getConnStatus() ConnStatus {
	return clientObj.connStatus
}

// 设置连接状态
func (clientObj *Client) setConnStatus(status ConnStatus) {
	clientObj.connStatus = status
}

// 追加发送的数据
// sendDataItemObj:待发送数据项
// priority:优先级
// 返回值：无
func (clientObj *Client) appendSendData(responseObj *serverResponseObject.ResponseObject, priority Priority) {
	clientObj.mutex.Lock()
	defer clientObj.mutex.Unlock()

	if priority == Con_LowPriority {
		clientObj.sendData_LowPriority = append(clientObj.sendData_LowPriority, responseObj)
	} else {
		clientObj.sendData = append(clientObj.sendData, responseObj)
	}
}

// 追加接收到的数据
// receiveData：接收到的数据
// 返回值：无
func (clientObj *Client) appendReceiveData(receiveData []byte) {
	clientObj.receiveData = append(clientObj.receiveData, receiveData...)
	clientObj.activeTime = time.Now()
}

// 发送字节数组消息
// responseObj:返回值对象
func (clientObj *Client) sendMessage(responseObj *serverResponseObject.ResponseObject) error {
	beforeTime := time.Now().Unix()

	//序列化发送的数据
	content, err := json.Marshal(responseObj)
	if err != nil {
		logUtil.Log("序列化response数据失败", logUtil.Error, true)
		return errors.New("序列化response数据失败")
	}

	// 获得数据内容的长度
	contentLength := len(content)

	// 将长度转化为字节数组
	header := intAndBytesUtil.Int32ToBytes(int32(contentLength), byterOrder)

	// 将头部与内容组合在一起
	message := append(header, content...)

	// 发送消息
	if _, err = clientObj.conn.Write(message); err != nil {
		logUtil.Log(fmt.Sprintf("发送消息,%s,出现错误：%s", content, err), logUtil.Error, true)
		return err
	}

	// 如果发送的时间超过3秒，则记录下来
	if time.Now().Unix()-beforeTime > 3 {
		logUtil.Log(fmt.Sprintf("消息Size:%d, UseTime:%d", contentLength, time.Now().Unix()-beforeTime), logUtil.Warn, true)
	}

	return err
}

// 判断客户端是否超时（超过300秒不活跃算作超时）
// 返回值：是否超时
func (c *Client) hasExpired() bool {
	return time.Now().Unix() > c.activeTime.Add(300*time.Second).Unix()
}

// 记录日志
// log：日志内容
func (clientObj *Client) WriteLog(log string) {
	if debug {
		fileUtil.WriteFile("Log", clientObj.getRemoteAddr(), true,
			timeUtil.Format(time.Now(), "yyyy-MM-dd HH:mm:ss"),
			" ",
			fmt.Sprintf("client:%s", clientObj.String()),
			" ",
			log,
			"\r\n",
			"\r\n",
		)
	}
}

// 格式化
func (clientObj *Client) String() string {
	return fmt.Sprintf("{Id:%d, RemoteAddr:%d, activeTime:%s, playerId:%s}", clientObj.id, clientObj.getRemoteAddr(), timeUtil.Format(clientObj.activeTime, "yyyy-MM-dd HH:mm:ss"), clientObj.playerId)
}

// 玩家登陆
// playerId：玩家Id
// 返回值：无
func (c *Client) PlayerLogin(playerId string) {
	c.playerId = playerId
}

// 玩家登出
// 返回值：无
func (c *Client) playerLogout() {
	c.playerId = ""
}

// 玩家登出，客户端退出
// 返回值：无
func (c *Client) LogoutAndQuit() {
	c.playerLogout()
	c.conn.Close()
	c.setConnStatus(con_Close)
}

// 新建客户端对象
// conn：连接对象
// 返回值：客户端对象的指针
func newClient(_conn net.Conn) *Client {
	// 获得自增的id值
	getIncrementId := func() int32 {
		atomic.AddInt32(&globalClientId, 1)
		return globalClientId
	}

	return &Client{
		id:                   getIncrementId(),
		conn:                 _conn,
		connStatus:           con_Open,
		receiveData:          make([]byte, 0, 1024),
		sendData:             make([]*serverResponseObject.ResponseObject, 0, 16),
		sendData_LowPriority: make([]*serverResponseObject.ResponseObject, 0, 16),
		activeTime:           time.Now(),
		playerId:             "",
	}
}
