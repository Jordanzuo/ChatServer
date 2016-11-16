package rpcServer

import (
	"fmt"
	"time"

	"github.com/Jordanzuo/goutil/logUtil"
)

// 清理过期的客户端
func clearExpiredClient() {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	for {
		// 休眠指定的时间（单位：秒）(放在此处是因为程序刚启动时并没有过期的客户端，所以先不用占用资源；并且此时LogPath尚未设置，如果直接执行后面的代码会出现panic异常)
		time.Sleep(5 * time.Minute)

		beforeClientCount := GetClientCount()
		beforePlayerCount := getPlayerCount()

		// 获取过期的客户端列表
		expiredClientList := getExpiredClientList()
		expiredClientCount := len(expiredClientList)
		if expiredClientCount == 0 {
			continue
		}

		for _, item := range expiredClientList {
			disconnectByClient(item)
		}

		// 记录日志
		logUtil.Log(fmt.Sprintf("清理前的客户端数量为：%d，清理前的玩家数量为：%d， 本次清理不活跃的数量为：%d", beforeClientCount, beforePlayerCount, expiredClientCount), logUtil.Debug, true)
	}
}
