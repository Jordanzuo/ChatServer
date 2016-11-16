package config

import (
	"github.com/Jordanzuo/goutil/configUtil"
	"github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/goutil/logUtil"
)

var (
	// 是否是DEBUG模式
	DEBUG bool

	// 数据库连接字符串
	DBConnection string

	// 聊天服务器监听地址
	ChatServerListenAddress string

	// 聊天服务器公网地址
	ChatServerPublicAddress string
)

func init() {
	// 设置日志文件的存储目录
	logUtil.SetLogPath("LOG")

	// 读取配置文件内容
	config, err := configUtil.ReadJsonConfig("config.ini")
	checkError(err)

	// 解析DEBUG配置
	debug, err := configUtil.ReadBoolJsonValue(config, "DEBUG")
	checkError(err)

	// 为DEBUG模式赋值
	DEBUG = debug

	// 设置debugUtil的状态
	debugUtil.SetDebug(debug)

	// 解析mysql配置数据
	DBConnection, err = configUtil.ReadStringJsonValue(config, "DBConnection")
	checkError(err)

	// 解析ChatServerListenAddress
	ChatServerListenAddress, err = configUtil.ReadStringJsonValue(config, "ChatServerListenAddress")
	checkError(err)

	// 解析ChatServerPublicAddress
	ChatServerPublicAddress, err = configUtil.ReadStringJsonValue(config, "ChatServerPublicAddress")
	checkError(err)

	debugUtil.Println("DEBUG:", debug)
	debugUtil.Println("DBConnection:", DBConnection)
	debugUtil.Println("ChatServerListenAddress:", ChatServerListenAddress)
	debugUtil.Println("ChatServerPublicAddress:", ChatServerPublicAddress)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
