package chatBLL

import (
	"fmt"
	"strconv"

	"github.com/Jordanzuo/ChatServer/src/bll/configBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/manageCenterBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/playerBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/wordBLL"
	"github.com/Jordanzuo/ChatServer/src/rpcClient"
	"github.com/Jordanzuo/ChatServer/src/rpcServer"
	"github.com/Jordanzuo/ChatServerModel/src/channelType"
	"github.com/Jordanzuo/ChatServerModel/src/commandType"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ChatServerModel/src/serverResponseObject"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/ManageCenterModel_Go/server"
	"github.com/Jordanzuo/ManageCenterModel_Go/serverGroup"
	_ "github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/goutil/securityUtil"
)

// 登陆
func Login(clientObj *rpcServer.Client, id, name, unionId, extraMsg, sign string, partnerId, serverId int) *serverResponseObject.ResponseObject {
	responseObj := serverResponseObject.NewResponseObject(commandType.Login)

	// 定义变量
	var err error
	var exists bool
	var isNewPlayer bool
	var playerObj *player.Player
	var gamePlayerName string
	var gameUnionId string
	var serverGroupObj *serverGroup.ServerGroup
	var serverObj *server.Server

	// 验证签名
	verifySign := func(id, name, sign string) bool {
		rawstring := fmt.Sprintf("%s-%s-%s", id, name, configBLL.GetConfig().GetAppKey())
		if sign == securityUtil.Md5String(rawstring, false) {
			return true
		}

		return false
	}

	// 验证签名是否正确
	if verifySign(id, name, sign) == false {
		return responseObj.SetResultStatus(serverResponseObject.Con_SignError)
	}

	// 判断服务器组是否存在
	if serverGroupObj, serverObj, exists = manageCenterBLL.GetServerGroup(partnerId, serverId); !exists {
		return responseObj.SetResultStatus(serverResponseObject.Con_ServerGroupNotExist)
	}

	// 判断玩家是否在缓存中已经存在
	playerObj, exists, err = playerBLL.GetPlayer(id, false)
	if err != nil {
		return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
	}

	if exists {
		// 判断是否重复登陆
		if playerObj.ClientId > 0 {
			if oldClientObj, exists := rpcServer.GetClient(playerObj.ClientId); exists {
				// 如果不是同一个客户端，则先给客户端发送在其他设备登陆信息，然后断开连接
				if clientObj != oldClientObj {
					playerBLL.SendLoginAnotherDeviceMsg(oldClientObj)
				}
			}
		}
	} else {
		// 判断数据库中是否已经存在该玩家，如果不存在则表明是新玩家，先到游戏中验证
		playerObj, exists, err = playerBLL.GetPlayer(id, true)
		if err != nil {
			return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
		}

		if !exists {
			// 验证玩家Id在游戏库中是否存在
			gamePlayerName, gameUnionId, _, exists, err = playerBLL.GetGamePlayer(serverGroupObj, id)
			if err != nil {
				return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
			} else if !exists {
				return responseObj.SetResultStatus(serverResponseObject.Con_PlayerNotExist)
			} else {
				if name != gamePlayerName {
					return responseObj.SetResultStatus(serverResponseObject.Con_NameError)
				}

				if !playerBLL.IsUnionIdEmpty(unionId) && unionId != gameUnionId {
					return responseObj.SetResultStatus(serverResponseObject.Con_UnionIdError)
				}
			}

			if playerObj, err = playerBLL.RegisterNewPlayer(id, name, partnerId, serverId, unionId, extraMsg); err != nil {
				return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
			}
			isNewPlayer = true
		}
	}

	// 判断玩家是否被封号
	if playerObj.IsForbidden {
		return responseObj.SetResultStatus(serverResponseObject.Con_PlayerIsForbidden)
	}

	// 更新客户端对象的玩家Id
	clientObj.PlayerLogin(id)

	// 更新玩家登录信息
	if err = playerBLL.UpdateLoginInfo(playerObj, clientObj, isNewPlayer); err != nil {
		return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
	}

	// 设置玩家的服务器信息
	playerObj.SetServerInfo(serverGroupObj.Id, serverObj.Name)

	// 将玩家对象添加到玩家列表中
	playerBLL.RegisterPlayer(playerObj)

	// 输出结果
	playerBLL.SendToClient(clientObj, responseObj)

	return responseObj
}

// 登出
func Logout(clientObj *rpcServer.Client, playerObj *player.Player) *serverResponseObject.ResponseObject {
	responseObj := serverResponseObject.NewResponseObject(commandType.Logout)

	// 玩家登出
	clientObj.LogoutAndQuit()

	// 将玩家对象从缓存中移除
	playerBLL.UnRegisterPlayer(playerObj)

	return responseObj
}

// 更新玩家信息
func UpdatePlayerInfo(clientObj *rpcServer.Client, playerObj *player.Player, name, unionId, extraMsg string) *serverResponseObject.ResponseObject {
	responseObj := serverResponseObject.NewResponseObject(commandType.UpdatePlayerInfo)

	// 定义变量
	var gamePlayerName string
	var gameUnionId string
	var exists bool
	var err error
	var serverGroupObj *serverGroup.ServerGroup

	// 如果玩家名或公会Id有改变，则到游戏库中去验证是否是正确的名称
	if name != playerObj.Name || unionId != playerObj.UnionId {
		// 判断服务器组是否存在
		if serverGroupObj, _, exists = manageCenterBLL.GetServerGroup(playerObj.PartnerId, playerObj.ServerId); !exists {
			return responseObj.SetResultStatus(serverResponseObject.Con_ServerGroupNotExist)
		}

		// 验证玩家Id在游戏库中是否存在
		gamePlayerName, gameUnionId, _, exists, err = playerBLL.GetGamePlayer(serverGroupObj, playerObj.Id)
		if err != nil {
			return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
		} else if !exists {
			return responseObj.SetResultStatus(serverResponseObject.Con_PlayerNotExist)
		} else {
			if name != gamePlayerName {
				return responseObj.SetResultStatus(serverResponseObject.Con_NameError)
			}

			if !playerBLL.IsUnionIdEmpty(unionId) && unionId != gameUnionId {
				return responseObj.SetResultStatus(serverResponseObject.Con_UnionIdError)
			}
		}
	}

	// 更新玩家信息
	// 判断是否有信息变化
	if name != playerObj.Name || unionId != playerObj.UnionId || extraMsg != playerObj.ExtraMsg {
		if err = playerBLL.UpdateInfo(playerObj, name, unionId, extraMsg); err != nil {
			return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
		}
	}

	// 输出结果
	playerBLL.SendToClient(clientObj, responseObj)

	return responseObj
}

// 发送消息
func SendMessage(clientObj *rpcServer.Client, playerObj *player.Player, _channelType channelType.ChannelType, message, toPlayerId string) *serverResponseObject.ResponseObject {
	responseObj := serverResponseObject.NewResponseObject(commandType.SendMessage)

	// 判断玩家是否被禁言
	if isInSilent, _ := playerObj.IsInSilent(); isInSilent {
		return responseObj.SetResultStatus(serverResponseObject.Con_PlayerIsInSilent)
	}

	switch _channelType {
	case channelType.World:
		// 如果是世界频道，则判断禁止词汇
		if wordBLL.IfContainsForbidWords(message) {
			return responseObj.SetResultStatus(serverResponseObject.Con_ContainForbiddenWord)
		}
	case channelType.Union:
		// 判断公会Id是否为空
		if playerBLL.IsUnionIdEmpty(playerObj.UnionId) {
			return responseObj.SetResultStatus(serverResponseObject.Con_NotInUnion)
		}
	case channelType.Private:
		// 目标玩家Id不能为空
		if toPlayerId == "" {
			return responseObj.SetResultStatus(serverResponseObject.Con_NotFoundTarget)
		}

		// 不能给自己发送消息
		if toPlayerId == playerObj.Id {
			return responseObj.SetResultStatus(serverResponseObject.Con_CantSendMessageToSelf)
		}
	case channelType.CrossServer:
		// 如果是世界频道，则判断禁止词汇
		if wordBLL.IfContainsForbidWords(message) {
			return responseObj.SetResultStatus(serverResponseObject.Con_ContainForbiddenWord)
		}

		// 判断是否可以向所有服务器发送信息
		// 判断服务器组是否存在
		if serverGroupObj, _, exists := manageCenterBLL.GetServerGroup(playerObj.PartnerId, playerObj.ServerId); !exists {
			return responseObj.SetResultStatus(serverResponseObject.Con_ServerGroupNotExist)
		} else {
			_, _, isCrossServer, exists, err := playerBLL.GetGamePlayer(serverGroupObj, playerObj.Id)
			if err != nil {
				return responseObj.SetResultStatus(serverResponseObject.Con_DataError)
			} else if !exists {
				return responseObj.SetResultStatus(serverResponseObject.Con_PlayerNotExist)
			} else if !isCrossServer {
				return responseObj.SetResultStatus(serverResponseObject.Con_CantSendCrossServerMessage)
			}
		}
	default:
		return responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
	}

	// debugUtil.Printf("playerObj:%v, ServerGroupId:%v\n", playerObj, playerObj.ServerGroupId)

	chatMessageObj := transferObject.NewChatMessageObject(_channelType, strconv.Itoa(playerObj.ServerGroupId), message, playerObj)
	chatMessageObj.SetToPlayerId(toPlayerId)
	rpcClient.ChatMessageObjectChannel <- chatMessageObj

	// debugUtil.Printf("chatMessageObj.Player:%v, ServerGroupId:%v\n", chatMessageObj, chatMessageObj.Player.ServerGroupId)

	return responseObj
}
