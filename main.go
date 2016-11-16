package main

import (
	_ "github.com/Jordanzuo/ChatServer/src/bll/wordBLL"
)

import (
	"fmt"
	"github.com/Jordanzuo/ChatServer/src/bll/chatBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/configBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/playerBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/reloadBLL"
	"github.com/Jordanzuo/ChatServer/src/config"
	"github.com/Jordanzuo/ChatServer/src/rpcClient"
	"github.com/Jordanzuo/ChatServer/src/rpcServer"
	"github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/goutil/logUtil"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

var (
	wg sync.WaitGroup
)

func init() {
	// 设置WaitGroup需要等待的数量，只要有一个服务器出现错误都停止服务器
	wg.Add(1)
}

// 处理系统信号
func signalProc() {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for {
		// 准备接收信息
		sig := <-sigs

		// 输出信号
		debugUtil.Println("sig:", sig)

		if sig == syscall.SIGHUP {
			logUtil.Log("收到重启的信号，准备重新加载配置", logUtil.Info, true)

			// 重新加载配置
			reloadBLL.Reload()

			logUtil.Log("收到重启的信号，重新加载配置完成", logUtil.Info, true)
		} else {
			logUtil.Log("收到退出程序的信号，开始退出……", logUtil.Info, true)

			// 做一些收尾的工作

			logUtil.Log("收到退出程序的信号，退出完成……", logUtil.Info, true)

			// 一旦收到信号，则表明管理员希望退出程序，则先保存信息，然后退出
			os.Exit(0)
		}
	}
}

// 记录当前运行的Goroutine数量
func recordGoroutineNum() {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	for {
		time.Sleep(5 * time.Minute)

		// 记录当前运行的Goroutine数量
		logUtil.Log(fmt.Sprintf("NumGoroutine:%d", runtime.NumGoroutine()), logUtil.Debug, true)
	}
}

func main() {
	// 处理系统信号
	go signalProc()

	// 记录当前运行的Goroutine数量
	go recordGoroutineNum()

	// 获取数据库配置
	configObj := configBLL.GetConfig()

	// 设置rpcClient配置，并启动服务器
	rpcClient.SetConfig(configObj.GetChatServerCenterRpcAddress(),
		config.ChatServerPublicAddress,
		chatBLL.HandleCenterMessage,
		rpcServer.GetClientCount,
		playerBLL.GetPlayerCount,
		config.DEBUG)
	rpcClient.StartClient(true)

	// 设置rpcServer配置，并启动服务器
	rpcServer.SetConfig(config.ChatServerListenAddress,
		playerBLL.GetPlayer,
		playerBLL.GetPlayerCount,
		playerBLL.DisconnectByClient,
		chatBLL.Login,
		chatBLL.Logout,
		chatBLL.UpdatePlayerInfo,
		chatBLL.SendMessage,
		config.DEBUG)
	go rpcServer.StartServer(&wg)

	// 阻塞等待，以免main线程退出
	wg.Wait()
}
