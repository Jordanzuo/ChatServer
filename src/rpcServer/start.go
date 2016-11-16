package rpcServer

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/Jordanzuo/goutil/logUtil"
)

// 启动服务器
//此处开启的goroutine不需要捕获异常
func StartServer(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	logUtil.Log("Socket服务器开始监听...", logUtil.Info, true)

	// 监听指定的端口
	listener, err := net.Listen("tcp", chatServerListenAddress)
	if err != nil {
		panic(errors.New(fmt.Sprintf("Listen Error: %s", err)))
	} else {
		msg := fmt.Sprintf("Got listener for the server. (local address: %s)", listener.Addr())

		// 记录和显示日志，并且判断是否需要退出
		logUtil.Log(msg, logUtil.Info, true)
		fmt.Println(msg)
	}

	// 清理过期的客户端
	go clearExpiredClient()

	// 显示数据大小信息(每5分钟更新一次)
	go displayDataSize()

	for {
		// 阻塞直至新连接到来
		conn, err := listener.Accept()
		if err != nil {
			logUtil.Log(fmt.Sprintf("Accept Error: %s", err), logUtil.Error, true)
			continue
		}

		// 启动一个新协程来处理链接
		go handleConn(conn)
	}
}
