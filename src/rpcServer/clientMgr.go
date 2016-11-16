package rpcServer

import (
	"sync"
)

var (
	// 客户端连接列表
	clientMap = make(map[int32]*Client, 1024)

	// 读写锁
	mutex sync.RWMutex
)

// 添加新的客户端
// clientObj：客户端对象
func registerClient(clientObj *Client) {
	mutex.Lock()
	defer mutex.Unlock()

	clientMap[clientObj.GetId()] = clientObj
}

// 移除客户端
// clientObj：客户端对象
func UnRegisterClient(clientObj *Client) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientMap, clientObj.GetId())
}

// 根据客户端Id获取对应的客户端对象
// id：客户端Id
// 返回值：客户端对象
func GetClient(id int32) (*Client, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	if clientObj, ok := clientMap[id]; ok {
		return clientObj, true
	}

	return nil, false
}

// 获取客户端的数量
// 返回值：
// 客户端数量
func GetClientCount() int {
	mutex.RLock()
	defer mutex.RUnlock()

	return len(clientMap)
}

// 返回过期的客户端列表
// 返回值：
// 过期的客户端列表
func getExpiredClientList() (expiredClientList []*Client) {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, item := range clientMap {
		if item.hasExpired() {
			expiredClientList = append(expiredClientList, item)
		}
	}

	return
}
