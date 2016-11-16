package rpcServer

import (
	"github.com/Jordanzuo/ChatServerModel/src/channelType"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ChatServerModel/src/serverResponseObject"
)

var (
	// 聊天服务器监听地址
	chatServerListenAddress string

	//查找player方法
	getPlayer func(string, bool) (*player.Player, bool, error)

	//查找getPlayerCount方法
	getPlayerCount func() int

	//查找disconnectByClient方法
	disconnectByClient func(*Client)

	// 登陆处理器
	loginHandler func(*Client, string, string, string, string, string, int, int) *serverResponseObject.ResponseObject

	// 登出处理器
	logoutHandler func(*Client, *player.Player) *serverResponseObject.ResponseObject

	// 更新玩家信息处理器
	updatePlayerInfoHandler func(*Client, *player.Player, string, string, string) *serverResponseObject.ResponseObject

	// 发送消息处理器
	sendMessageHandler func(*Client, *player.Player, channelType.ChannelType, string, string) *serverResponseObject.ResponseObject

	// 是否测试
	debug bool
)

//传递上层函数地址
func SetConfig(_chatServerListenAddress string,
	_getPlayer func(string, bool) (*player.Player, bool, error),
	_getPlayerCount func() int,
	_disconnectByClient func(*Client),
	_loginHandler func(*Client, string, string, string, string, string, int, int) *serverResponseObject.ResponseObject,
	_logoutHandler func(*Client, *player.Player) *serverResponseObject.ResponseObject,
	_updatePlayerInfoHandler func(*Client, *player.Player, string, string, string) *serverResponseObject.ResponseObject,
	_sendMessageHandler func(*Client, *player.Player, channelType.ChannelType, string, string) *serverResponseObject.ResponseObject,
	_debug bool) {

	// 为配置赋值
	chatServerListenAddress = _chatServerListenAddress
	getPlayer = _getPlayer
	getPlayerCount = _getPlayerCount
	disconnectByClient = _disconnectByClient
	loginHandler = _loginHandler
	logoutHandler = _logoutHandler
	updatePlayerInfoHandler = _updatePlayerInfoHandler
	sendMessageHandler = _sendMessageHandler
	debug = _debug
}
