package chatBLL

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Jordanzuo/ChatServer/src/bll/manageCenterBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/playerBLL"
	"github.com/Jordanzuo/ChatServer/src/bll/reloadBLL"
	"github.com/Jordanzuo/ChatServerModel/src/channelType"
	"github.com/Jordanzuo/ChatServerModel/src/commandType"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ChatServerModel/src/playerDisconnectType"
	"github.com/Jordanzuo/ChatServerModel/src/serverResponseData"
	"github.com/Jordanzuo/ChatServerModel/src/serverResponseObject"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/ManageCenterModel_Go/serverGroup"
	"github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/goutil/logUtil"
)

// 处理消息（来自于ChatServerCenter的消息）
func HandleCenterMessage(forwardObj *transferObject.ForwardObject) {
	// 设置消息到达服务器的时间
	// forwardObj.ChatMessageObject.ArriveServerTime = time.Now()

	// 根据TransferType来选择数据
	switch forwardObj.MessageType {
	case transferObject.ChatMessage:
		handleChatMessage(forwardObj.ChatMessageObject)
	case transferObject.PushMessage:
		handlePushMessage(forwardObj.ChatMessageObject)
	case transferObject.Forbid:
		handleForbid(forwardObj.ChatMessageObject)
	case transferObject.Silent:
		handleSilent(forwardObj.ChatMessageObject)
	case transferObject.Reload:
		handleReload(forwardObj.ChatMessageObject)
	default:
		logUtil.Log(fmt.Sprintf("从Center收到了未定义的类型%s", forwardObj.MessageType), logUtil.Error, true)
	}
}

func handleChatMessage(chatMessageObj *transferObject.ChatMessageObject) {
	responseObj := serverResponseObject.NewResponseObject(commandType.SendMessage)

	// 定义变量
	var finalPlayerList = make([]*player.Player, 0, 1024)
	var toPlayerObj *player.Player
	var ifToPlayerExists bool
	var err error

	// debugUtil.Printf("chatMessageObj.ChannelType:%v, chatMessageObj.Player:%v\n", chatMessageObj.ChannelType, chatMessageObj.Player)

	switch chatMessageObj.ChannelType {
	case channelType.World:
		finalPlayerList = playerBLL.GetPlayerListInSameServerGroup(chatMessageObj.Player)
	case channelType.Union:
		finalPlayerList = playerBLL.GetPlayerListInSameUnion(chatMessageObj.Player)
	case channelType.Private:
		// 获得目标玩家对象
		toPlayerObj, ifToPlayerExists, err = playerBLL.GetPlayer(chatMessageObj.ToPlayerId, false)
		if err != nil {
			return
		}

		if !ifToPlayerExists {
			return
		}

		// 判断目标玩家是否在同一区服
		var selfServerGroupObj *serverGroup.ServerGroup
		var toServerGroupObj *serverGroup.ServerGroup
		var exists bool
		if selfServerGroupObj, _, exists = manageCenterBLL.GetServerGroup(chatMessageObj.Player.PartnerId, chatMessageObj.Player.ServerId); !exists {
			return
		}
		if toServerGroupObj, _, exists = manageCenterBLL.GetServerGroup(toPlayerObj.PartnerId, toPlayerObj.ServerId); !exists {
			return
		}
		if selfServerGroupObj != toServerGroupObj {
			return
		}

		// 添加到列表中
		finalPlayerList = append(finalPlayerList, chatMessageObj.Player, toPlayerObj)
	case channelType.CrossServer:
		finalPlayerList = playerBLL.GetAllPlayerList()
	default:
		return
	}

	debugUtil.Printf("finalPlayerList:%v\n", finalPlayerList)

	// 设置responseObj的Data属性
	responseObj.SetData(serverResponseData.NewResponseData(chatMessageObj.ChannelType, chatMessageObj.Message, chatMessageObj.Player, toPlayerObj))

	// 向玩家发送消息
	playerBLL.SendToPlayer(finalPlayerList, responseObj)
}

func handlePushMessage(chatMessageObj *transferObject.ChatMessageObject) {
	responseObj := serverResponseObject.NewResponseObject(commandType.SendMessage)

	finalPlayerList := make([]*player.Player, 0, 32)
	if chatMessageObj.ToPlayerIds != nil && len(chatMessageObj.ToPlayerIds) > 0 {
		for _, playerId := range chatMessageObj.ToPlayerIds {
			if playerObj, exists, err := playerBLL.GetPlayer(playerId, false); err == nil && exists {
				finalPlayerList = append(finalPlayerList, playerObj)
			}
		}
	} else if chatMessageObj.ToUnionId != "" {
		// debugUtil.Printf("ServerGroupIds:%v\n", chatMessageObj.ServerGroupIds)
		if serverGroupId, err := strconv.Atoi(chatMessageObj.ServerGroupIds); err == nil {
			// debugUtil.Printf("serverGroupId:%v\n", serverGroupId)
			if _, exists := manageCenterBLL.GetServerGroupItem(serverGroupId); exists {
				finalPlayerList = append(finalPlayerList, playerBLL.GetPlayerListInUnion(serverGroupId, chatMessageObj.ToUnionId)...)
			}
		}
	} else {
		if chatMessageObj.ServerGroupIds == "0" {
			finalPlayerList = playerBLL.GetAllPlayerList()
		} else {
			// debugUtil.Printf("ServerGroupIds:%s\n", chatMessageObj.ServerGroupIds)
			for _, serverGroupId_str := range strings.Split(chatMessageObj.ServerGroupIds, ",") {
				if serverGroupId, err := strconv.Atoi(serverGroupId_str); err == nil {
					// debugUtil.Printf("ServerGroupId:%d\n", serverGroupId)
					if _, exists := manageCenterBLL.GetServerGroupItem(serverGroupId); exists {
						finalPlayerList = append(finalPlayerList, playerBLL.GetPlayerListInServerGroup(serverGroupId)...)
					}
				}
			}
		}
	}

	// 设置responseObj的Data属性
	responseObj.SetData(serverResponseData.NewResponseData(chatMessageObj.ChannelType, chatMessageObj.Message, nil, nil))

	// 向玩家发送消息
	playerBLL.SendToPlayer(finalPlayerList, responseObj)

	// 设置服务器处理消息结束的时间
	// chatMessageObj.HandleEndTime = time.Now()
	// logUtil.Log(fmt.Sprintf("chatMessageObj:%v", chatMessageObj), logUtil.Debug, true)
}

func handleForbid(chatMessageObj *transferObject.ChatMessageObject) {
	// 判断玩家对象是否存在
	playerObj, exists, err := playerBLL.GetPlayer(chatMessageObj.ForbidPlayerId, false)
	if err != nil || exists == false {
		// debugUtil.Printf("找不到被封号的玩家%s\n", chatMessageObj.ForbidPlayerId)
		return
	}

	// 断开客户端连接
	playerBLL.DisconnectByPlayer(playerObj, playerDisconnectType.Con_FromForbid)
}

func handleSilent(chatMessageObj *transferObject.ChatMessageObject) {
	// 判断玩家对象是否存在
	playerObj, exists, err := playerBLL.GetPlayer(chatMessageObj.SilentPlayerId, false)
	if err != nil || exists == false {
		// debugUtil.Printf("找不到被禁言的玩家%s\n", chatMessageObj.SilentPlayerId)
		return
	}

	// 更新禁言结束时间
	playerObj.SilentEndTime = chatMessageObj.SilentEndTime
	// debugUtil.Printf("SilentEndTime:%v\n", playerObj.SilentEndTime)
}

func handleReload(chatMessageObj *transferObject.ChatMessageObject) {
	// reload
	errList := reloadBLL.Reload()
	if errList != nil && len(errList) > 0 {
		for _, err := range errList {
			logUtil.Log(fmt.Sprintf("Reload Err:%s", err), logUtil.Error, true)
		}
	}
}
