package rpcClient

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Jordanzuo/ChatServerModel/src/centerRequestObject"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/goutil/debugUtil"
	"github.com/Jordanzuo/goutil/logUtil"
)

var (
	// 请求Id:每个请求都会带上一个唯一Id，以便在接收到服务器的返回数据时能够区分出来自于不同的请求
	requestId int32 = 0

	// 回调方法集合，及其锁对象
	callbackFuncMap = make(map[int32]func(interface{}))
	mutex           sync.Mutex
)

// 注册回调方法
// id:自增Id
// callback:回调方法
func registerCallbackFunc(id int32, callbackFunc func(interface{})) {
	mutex.Lock()
	defer mutex.Unlock()

	callbackFuncMap[id] = callbackFunc
}

// 获取回调方法
// id:自增Id
// 返回值：
// 回调方法
func getCallbackFunc(id int32) (callbackFunc func(interface{}), exists bool) {
	mutex.Lock()
	defer mutex.Unlock()

	callbackFunc, exists = callbackFuncMap[id]

	return
}

// 删除回调方法
// id:自增Id
func deleteCallbackFunc(id int32) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(callbackFuncMap, id)
}

// 向服务端发送请求
// transferType：传输类型
// parameters：调用的方法参数
// function：请求对应的回调方法
func request(transferType transferObject.TransferType, parameters []interface{}, function func(interface{})) {
	getIncrementmId := func() int32 {
		atomic.AddInt32(&requestId, 1)
		return requestId
	}

	id := getIncrementmId()
	requestObj := centerRequestObject.NewRequestObject(string(transferType), parameters)

	if b, err := json.Marshal(requestObj); err != nil {
		logUtil.Log(fmt.Sprintf("序列化请求数据%v出错", requestObj), logUtil.Error, true)
	} else {
		// 注册回调方法
		registerCallbackFunc(id, function)

		// 发送数据
		if clientObj := getClientObj(); clientObj != nil {
			clientObj.sendByteMessage(id, b)
			debugUtil.Println("request id is:", id)
		}
	}
}
