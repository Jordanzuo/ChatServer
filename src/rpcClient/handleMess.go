package rpcClient

import (
	"github.com/Jordanzuo/ChatServerModel/src/centerResponseObject"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/goutil/debugUtil"
)

// 处理由服务器主动推送过来的消息
func handleActiveMess(forwardObj *transferObject.ForwardObject) {
	// 由HandleCenterMessage方法来进行处理
	go handleCenterMessage(forwardObj)
}

// 处理由客户端发送给服务器,再由服务器反馈的消息
func handlePassiveMess(id int32, responseObj *centerResponseObject.ResponseObject) {
	callbackFunc, exists := getCallbackFunc(id)
	if !exists {
		debugUtil.Println("receive response is invalid data, id is :", id)
		return
	}

	defer func() {
		deleteCallbackFunc(id)
	}()

	// 返回成功，则调用指定的回调方法；否则表示一些提示、警告、或者版本、资源更新等信息；否则表示其它信息的返回
	if responseObj.Code == centerResponseObject.Con_Success {
		if callbackFunc == nil {
			debugUtil.Println("receive response from server success,but callbackFunc is nil in local")
			return
		}

		//调用对应的回调函数
		callbackFunc(responseObj.Data)
	} else {
		// 处理特殊的返回值
		switch responseObj.Code {
		default:
			debugUtil.Println("receive response from server failed，the error info is：", responseObj.Message)
		}
	}
}
