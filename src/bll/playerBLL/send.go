package playerBLL

import (
	"github.com/Jordanzuo/ChatServer/src/rpcServer"
	"github.com/Jordanzuo/ChatServerModel/src/commandType"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ChatServerModel/src/serverResponseObject"
	"time"
)

// 发送在另一台设备登陆的信息
// clientObj：客户端对象
func SendLoginAnotherDeviceMsg(clientObj *rpcServer.Client) {
	responseObj := serverResponseObject.NewResponseObject(commandType.Login)
	responseObj.SetResultStatus(serverResponseObject.Con_LoginOnAnotherDevice)

	// 先发送消息，然后再断开连接
	rpcServer.ResponseResult(clientObj, responseObj, rpcServer.Con_HighPriority)

	// 启动独立goroutine来发断开连接
	go func() {
		time.Sleep(2 * time.Second)
		clientObj.LogoutAndQuit()
	}()
}

// 发送封号信息
// clientObj：客户端对象
func SendForbidMsg(clientObj *rpcServer.Client) {
	responseObj := serverResponseObject.NewResponseObject(commandType.Login)
	responseObj.SetResultStatus(serverResponseObject.Con_PlayerIsForbidden)

	// 先发送消息，然后再断开连接
	rpcServer.ResponseResult(clientObj, responseObj, rpcServer.Con_HighPriority)

	// 启动独立goroutine来发断开连接
	go func() {
		time.Sleep(2 * time.Second)
		clientObj.LogoutAndQuit()
	}()
}

// 发送数据给客户端
// player：玩家对象
// responseObj：Socket服务器的返回对象
func SendToClient(clientObj *rpcServer.Client, responseObj *serverResponseObject.ResponseObject) {
	rpcServer.ResponseResult(clientObj, responseObj, rpcServer.Con_HighPriority)
}

// 发送数据给玩家
// playerList：玩家列表
// responseObj：Socket服务器的返回对象
func SendToPlayer(playerList []*player.Player, responseObj *serverResponseObject.ResponseObject) {
	for _, item := range playerList {
		if item.ClientId > 0 {
			if clientObj, ok := rpcServer.GetClient(item.ClientId); ok {
				rpcServer.ResponseResult(clientObj, responseObj, rpcServer.Con_HighPriority)
			}
		}
	}
}
