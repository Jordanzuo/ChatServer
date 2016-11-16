package rpcClient

import (
	"encoding/json"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/goutil/logUtil"
	"time"
)

var (
	ChatMessageObjectChannel = make(chan *transferObject.ChatMessageObject, 1024*100)
)

func init() {
	go func() {
		// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
		defer func() {
			if r := recover(); r != nil {
				logUtil.LogUnknownError(r)
			}
		}()

		for {
			select {
			case chatMessageObj := <-ChatMessageObjectChannel:
				go forward(chatMessageObj)
			default:
				// 如果channel中没有数据，则休眠5毫秒
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()
}

// 转发聊天消息
func forward(chatMessageObj *transferObject.ChatMessageObject) {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	if message, err := json.Marshal(chatMessageObj); err == nil {
		params := make([]interface{}, 1, 1)
		params[0] = string(message)

		//发送请求
		request(transferObject.Forward, params, nil)
	}
}
