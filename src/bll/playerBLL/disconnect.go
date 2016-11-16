package playerBLL

import (
	"github.com/Jordanzuo/ChatServer/src/rpcServer"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ChatServerModel/src/playerDisconnectType"
)

// 根据客户端对象来断开连接
// 注销客户端连接
// 从缓存中移除玩家对象
// clientObj：客户端对象
// clinetDisconnectType：客户端断开连接的类型；如果是来自于rpc则意味着之前客户端已经关闭连接，现在需要将客户端对象从缓存中移除了；否则是客户端过期，需要关闭
func DisconnectByClient(clientObj *rpcServer.Client) {
	// 将玩家从缓存中移除
	if clientObj.GetPlayerId() != "" {
		if playerObj, exists, err := GetPlayer(clientObj.GetPlayerId(), false); err == nil && exists {
			UnRegisterPlayer(playerObj)
		}
	}

	// 注销客户端连接，并从缓存中移除
	clientObj.LogoutAndQuit()
	rpcServer.UnRegisterClient(clientObj)
}

// 根据玩家对象来断开客连接
// 注销客户端连接
// 从缓存中移除玩家对象
// playerObj：玩家对象
// playerDisconnectType：玩家断开连接的类型
func DisconnectByPlayer(playerObj *player.Player, _playerDisconnectType playerDisconnectType.PlayerDisconnectType) {
	// 断开客户端连接
	if playerObj.ClientId > 0 {
		if clientObj, ok := rpcServer.GetClient(playerObj.ClientId); ok {
			switch _playerDisconnectType {
			case playerDisconnectType.Con_FromForbid:
				SendForbidMsg(clientObj)
			}
		}
	}

	// 将玩家从缓存中移除
	UnRegisterPlayer(playerObj)
}
