package rpcClient

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/Jordanzuo/ChatServerModel/src/centerResponseObject"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/goutil/logUtil"

	"time"
)

var (
	// 客户端对象
	clientObj *client

	// 存储登陆成功信息的通道
	loginSucceedCh = make(chan int)
)

func init() {
	// 保证与ChatServerCenter的连接
	go func() {
		// 处理内部未处理的异常，避免导致系统崩溃
		defer func() {
			if r := recover(); r != nil {
				logUtil.LogUnknownError(r)
			}
		}()

		for {
			// 先休眠5s
			time.Sleep(5 * time.Second)

			if clientObj == nil && chatServerCenterRpcAddress != "" && chatServerPublicAddress != "" {
				logUtil.Log("与ChatServerCenter的连接已经断开，尝试重连", logUtil.Debug, true)
				StartClient(false)

				if clientObj != nil {
					logUtil.Log("与ChatServerCenter重连成功", logUtil.Debug, true)
				} else {
					logUtil.Log("与ChatServerCenter重连失败", logUtil.Debug, true)
				}
			}
		}
	}()
}

// 获取客户端连接对象
func getClientObj() *client {
	if clientObj == nil {
		logUtil.Log("ClientObj尚未设置或已经断开连接", logUtil.Error, true)
	}

	return clientObj
}

func heartBeat() {
	// 处理内部未处理的异常，避免导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	for {
		// 由于连接刚刚建立，所以无需发心跳包；等待一段时间之后再发
		time.Sleep(30 * time.Second)

		// 发送客户端与玩家数据更新
		updateClientAndPlayer()
	}
}

func callback(id int32, centerResponseData []byte) {
	// 如果id=0表示是服务器主动推送过来的消息，否则是客户端请求后的信息返回
	if id == 0 {
		// 将返回结果反序列化
		forwardObj := new(transferObject.ForwardObject)
		if err := json.Unmarshal(centerResponseData, forwardObj); err != nil {
			logUtil.Log(fmt.Sprintf("反序列化%s出错，错误信息为：%s", string(centerResponseData), err), logUtil.Error, true)
			return
		}

		handleActiveMess(forwardObj)
	} else {
		// 将返回结果反序列化
		responseObj := new(centerResponseObject.ResponseObject)
		if err := json.Unmarshal(centerResponseData, responseObj); err != nil {
			logUtil.Log(fmt.Sprintf("反序列化%s出错，错误信息为：%s", string(centerResponseData), err), logUtil.Error, true)
			return
		}

		handlePassiveMess(id, responseObj)
	}
}

func handleClient(clientObj *client) {
	for {
		id, content, ok := clientObj.getValidMessage()
		if !ok {
			break
		}

		// 处理数据，如果长度为0则表示心跳包
		if len(content) == 0 {
			continue
		} else {
			callback(id, content)
		}
	}
}

// 启动客户端
// ch：通道，用于传输连接成功的结果
func start(ch chan int) {
	// 连接指定的端口
	msg := ""
	conn, err := net.DialTimeout("tcp", chatServerCenterRpcAddress, 2*time.Second)
	if err != nil {
		msg = fmt.Sprintf("Dial Error: %s", err)
	} else {
		msg = fmt.Sprintf("Connect to the server. (local address: %s)", conn.LocalAddr())
	}

	logUtil.Log(msg, logUtil.Info, true)
	debugUtil.Println(msg)

	// 发送连接失败的通知
	if err != nil {
		ch <- 0
		return
	}

	// 创建客户端对象
	clientObj = newClient(conn)

	// 发送连接成功的通知
	ch <- 1

	defer func() {
		conn.Close()
		clientObj = nil
	}()

	// 死循环，不断地读取数据，解析数据，发送数据
	for {
		// 先读取数据，每次读取1024个字节
		readBytes := make([]byte, 1024)

		// Read方法会阻塞，所以不用考虑异步的方式
		n, err := conn.Read(readBytes)
		if err != nil {
			var errMsg string

			// 判断是连接关闭错误，还是普通错误
			if err == io.EOF {
				errMsg = fmt.Sprintf("另一端关闭了连接：%s，读取到的字节数为：%d", err, n)
				clientObj.conn.Close()
			} else {
				errMsg = fmt.Sprintf("读取数据错误：%s，读取到的字节数为：%d", err, n)
			}

			logUtil.Log(errMsg, logUtil.Error, true)

			//退出
			break
		}

		// 将读取到的数据追加到已获得的数据的末尾
		clientObj.appendContent(readBytes[:n])

		// 已经包含有效的数据，处理该数据
		handleClient(clientObj)
	}
}

// 启动客户端（连接ChatServerCenter）
// ifStart：是否为启动程序调用
func StartClient(ifStart bool) {
	// 监听连接成功通道
	ch := make(chan int)
	go start(ch)

	//阻塞直到连接成功
	ret := <-ch
	if ret == 0 {
		if ifStart {
			panic("连接ChatServerCenter失败，请检查配置")
		} else {
			return
		}
	}

	// 发送login消息
	login()

	//阻塞直到登录成功或超时
	select {
	case <-loginSucceedCh:
		// 发送心跳包
		go heartBeat()
	case <-func() chan bool {
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(30 * time.Second)
			timeout <- false
		}()
		return timeout
	}():
		debugUtil.Println("Login Timeout")

		// 如果登录失败，则将对象置空，以便下一次重新初始化
		clientObj = nil

		// 如果是启动程序调用，则panic，否则不处理
		if ifStart {
			panic("登录ChatServerCenter超时，请检查配置")
		}
	}
}
