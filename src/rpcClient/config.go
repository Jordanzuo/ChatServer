package rpcClient

import (
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
)

var (
	// 聊天中心服务器的地址
	chatServerCenterRpcAddress string

	// 聊天服务器公网地址
	chatServerPublicAddress string

	// 处理中心发来的数据
	handleCenterMessage func(*transferObject.ForwardObject)

	// 获取客户端数量的方法
	getClientCount func() int

	//获取玩家数量的方法
	getPlayerCount func() int

	// DEBUG
	debug bool
)

func SetConfig(_chatServerCenterRpcAddress, _chatServerPublicAddress string,
	_handleCenterMessage func(*transferObject.ForwardObject),
	_getClientCount func() int,
	_getPlayerCount func() int,
	_debug bool) {
	chatServerCenterRpcAddress = _chatServerCenterRpcAddress
	chatServerPublicAddress = _chatServerPublicAddress
	handleCenterMessage = _handleCenterMessage
	getClientCount = _getClientCount
	getPlayerCount = _getPlayerCount
	debug = _debug
}
