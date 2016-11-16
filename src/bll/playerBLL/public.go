package playerBLL

import (
	"time"

	"github.com/Jordanzuo/ChatServer/src/dal/playerDAL"
	"github.com/Jordanzuo/ChatServer/src/rpcServer"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/goutil/debugUtil"
)

// 注册新玩家
// id：玩家Id
// name：玩家名称
// partnerId：合作商Id
// serverId：服务器Id
// unionId：玩家公会Id
// extraMsg：玩家透传信息
// 返回值：
// 玩家对象
// 错误对象
func RegisterNewPlayer(id, name string, partnerId, serverId int, unionId, extraMsg string) (*player.Player, error) {
	playerObj := player.InitPlayer(id, name, partnerId, serverId, unionId, extraMsg)
	if err := playerDAL.Insert(playerObj); err != nil {
		return nil, err
	}

	return playerObj, nil
}

// 更新玩家信息
// playerObj：玩家对象
// name：玩家名称
// unionId：玩家公会Id
// extraMsg：玩家透传信息
func UpdateInfo(playerObj *player.Player, name, unionId, extraMsg string) error {
	playerObj.Name = name
	playerObj.UnionId = unionId
	playerObj.ExtraMsg = extraMsg

	return playerDAL.UpdateInfo(playerObj)
}

// 更新登录信息
// playerObj：玩家对象
// clientObj：客户端对象
// isNewPlayer：是否是新玩家
func UpdateLoginInfo(playerObj *player.Player, clientObj *rpcServer.Client, isNewPlayer bool) error {
	playerObj.ClientId = clientObj.GetId()
	playerObj.LoginTime = time.Now()

	debugUtil.Printf("isNewPlayer:%v\n", isNewPlayer)

	// 如果不是新玩家则更新登录时间，否则使用创建时指定的登录时间
	if !isNewPlayer {
		if err := playerDAL.UpdateLoginTime(playerObj); err != nil {
			return err
		}
	}

	return nil
}
