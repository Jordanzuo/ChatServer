package playerBLL

import (
	"fmt"
	// "github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/ChatServer/src/bll/manageCenterBLL"
	"github.com/Jordanzuo/ChatServer/src/dal/playerDAL"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ManageCenterModel_Go/serverGroup"
	"github.com/Jordanzuo/goutil/logUtil"
	"sync"
)

var (
	// 玩家集合
	playerMap   = make(map[string]*player.Player, 1024)
	playerMutex sync.RWMutex

	// 区服玩家集合
	serverGroupPlayerMap   = make(map[int]*player.ServerGroupPlayer)
	serverGroupPlayerMutex sync.RWMutex
)

func init() {
	// 先初始化服务器组玩家列表
	serverGroupMap := manageCenterBLL.GetServerGroupMap()
	initServerGroupPlayer(serverGroupMap)

	// 再注册通知事件方法
	manageCenterBLL.RegisterServerGroupChangeFunc("InitServerGroupPlayer", initServerGroupPlayer)
}

// 初始化服务器组对应的玩家列表
func initServerGroupPlayer(serverGroupMap map[int]*serverGroup.ServerGroup) {
	serverGroupPlayerMutex.Lock()
	defer serverGroupPlayerMutex.Unlock()

	for serverGroupId, _ := range serverGroupMap {
		if _, exists := serverGroupPlayerMap[serverGroupId]; !exists {
			serverGroupPlayerMap[serverGroupId] = player.NewServerGroupPlayer(serverGroupId)
		}
	}

	logUtil.Log("初始化服务器组对应的玩家列表", logUtil.Debug, true)
}

// 注册玩家对象到缓存中
// playerObj：玩家对象
func RegisterPlayer(playerObj *player.Player) {
	// 添加到玩家集合中
	playerMutex.Lock()
	defer playerMutex.Unlock()
	playerMap[playerObj.Id] = playerObj

	// 添加到区服玩家集合中
	serverGroupPlayerMutex.RLock()
	defer serverGroupPlayerMutex.RUnlock()

	if serverGroupPlayerObj, exists := serverGroupPlayerMap[playerObj.ServerGroupId]; exists {
		serverGroupPlayerObj.AddPlayer(playerObj)
	} else {
		logUtil.Log(fmt.Sprintf("未找到ServerGroupId:%d的ServerGroupPlayer对象", playerObj.ServerGroupId), logUtil.Error, true)
	}

	// debugUtil.Printf("playerMap:%v\n", playerMap)
	// debugUtil.Printf("serverGroupPlayerMap:%v\n", serverGroupPlayerMap)
}

// 从缓存中取消玩家注册
// playerObj：玩家对象
func UnRegisterPlayer(playerObj *player.Player) {
	// 从玩家集合中删除
	playerMutex.Lock()
	defer playerMutex.Unlock()
	delete(playerMap, playerObj.Id)

	// 从区服玩家集合中删除
	serverGroupPlayerMutex.RLock()
	defer serverGroupPlayerMutex.RUnlock()

	if serverGroupPlayerObj, exists := serverGroupPlayerMap[playerObj.ServerGroupId]; exists {
		serverGroupPlayerObj.DeletePlayer(playerObj)
	} else {
		logUtil.Log(fmt.Sprintf("未找到ServerGroupId:%d的ServerGroupPlayer对象", playerObj.ServerGroupId), logUtil.Error, true)
	}
}

// 根据Id获取玩家对象（先从缓存中取，取不到再从数据库中去取）
// id：玩家Id
// isLoadFromDB：是否要从数据库中获取数据
// 返回值：
// 玩家对象
// 是否存在该玩家
// 错误对象
func GetPlayer(id string, isLoadFromDB bool) (playerObj *player.Player, exists bool, err error) {
	if id == "" {
		return
	}

	getPlayerFromCache := func(_id string) (_playerObj *player.Player, _exists bool) {
		playerMutex.RLock()
		defer playerMutex.RUnlock()

		_playerObj, _exists = playerMap[id]
		return
	}

	getPlayerFromDB := func(_id string) (_playerObj *player.Player, _exists bool, _err error) {
		_playerObj, _exists, _err = playerDAL.GetPlayer(id)
		return
	}

	if isLoadFromDB {
		playerObj, exists, err = getPlayerFromDB(id)
	} else {
		playerObj, exists = getPlayerFromCache(id)
	}

	return
}

// 获取玩家数量
// 返回值：
// 玩家数量
func GetPlayerCount() int {
	playerMutex.RLock()
	defer playerMutex.RUnlock()

	return len(playerMap)
}

// 获取所有的玩家列表
// 返回值：
// 所有的玩家列表
func GetAllPlayerList() (finalPlayerList []*player.Player) {
	playerMutex.RLock()
	defer playerMutex.RUnlock()

	for _, item := range playerMap {
		finalPlayerList = append(finalPlayerList, item)
	}

	return
}

// 获取指定公会的所有玩家列表
// serverGroupId：服务器组Id
// unionId：公会Id
// 返回值：
// 指定公会的所有玩家列表
func GetPlayerListInUnion(serverGroupId int, unionId string) (finalPlayerList []*player.Player) {
	serverGroupPlayerMutex.RLock()
	defer serverGroupPlayerMutex.RUnlock()

	if serverGroupPlayerObj, exists := serverGroupPlayerMap[serverGroupId]; exists {
		finalPlayerList = append(finalPlayerList, serverGroupPlayerObj.GetPlayerListInUnion(unionId)...)
	} else {
		logUtil.Log(fmt.Sprintf("未找到ServerGroupId:%d的ServerGroupPlayer对象", serverGroupId), logUtil.Error, true)
	}

	return
}

// 获取指定玩家同工会的所有玩家列表
// playerObj：指定玩家
// 返回值：
// 同工会的所有玩家列表
func GetPlayerListInSameUnion(playerObj *player.Player) (finalPlayerList []*player.Player) {
	// 从区服玩家集合中删除
	serverGroupPlayerMutex.RLock()
	defer serverGroupPlayerMutex.RUnlock()

	if serverGroupPlayerObj, exists := serverGroupPlayerMap[playerObj.ServerGroupId]; exists {
		finalPlayerList = append(finalPlayerList, serverGroupPlayerObj.GetPlayerListInUnion(playerObj.UnionId)...)
	} else {
		logUtil.Log(fmt.Sprintf("未找到ServerGroupId:%d的ServerGroupPlayer对象", playerObj.ServerGroupId), logUtil.Error, true)
	}

	return
}

// 获取指定服务器组的所有玩家列表
// serverGroupId：服务器组Id
// 返回值：
// 指定公会的所有玩家列表
func GetPlayerListInServerGroup(serverGroupId int) (finalPlayerList []*player.Player) {
	serverGroupPlayerMutex.RLock()
	defer serverGroupPlayerMutex.RUnlock()

	if serverGroupPlayerObj, exists := serverGroupPlayerMap[serverGroupId]; exists {
		finalPlayerList = append(finalPlayerList, serverGroupPlayerObj.GetPlayerList()...)
	} else {
		logUtil.Log(fmt.Sprintf("未找到ServerGroupId:%d的ServerGroupPlayer对象", serverGroupId), logUtil.Error, true)
	}

	return
}

// 获取同服的所有玩家列表
// playerObj：指定玩家
// 返回值：
// 同服的所有玩家列表
func GetPlayerListInSameServerGroup(playerObj *player.Player) (finalPlayerList []*player.Player) {
	serverGroupPlayerMutex.RLock()
	defer serverGroupPlayerMutex.RUnlock()

	if serverGroupPlayerObj, exists := serverGroupPlayerMap[playerObj.ServerGroupId]; exists {
		finalPlayerList = append(finalPlayerList, serverGroupPlayerObj.GetPlayerList()...)
	} else {
		logUtil.Log(fmt.Sprintf("未找到ServerGroupId:%d的ServerGroupPlayer对象", playerObj.ServerGroupId), logUtil.Error, true)
	}

	return
}

// 判断工会ID是否为空
// unionId: 工会ID
// 返回值：BOOL
func IsUnionIdEmpty(unionId string) bool {
	if unionId == "" || unionId == "00000000-0000-0000-0000-000000000000" {
		return true
	}

	return false
}
