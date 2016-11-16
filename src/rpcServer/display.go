package rpcServer

import (
	"fmt"
	"time"

	"github.com/Jordanzuo/goutil/logUtil"
)

// 显示数据大小信息(每5分钟更新一次)
func displayDataSize() {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	for {
		// 刚启动时不需要显示信息，故将Sleep放在前面，而不是最后
		time.Sleep(time.Minute)

		// 组装需要记录的信息
		logUtil.Log(fmt.Sprintf("当前客户端数量：%d, 玩家数量：%d", GetClientCount(), getPlayerCount()), logUtil.Debug, true)
	}
}
