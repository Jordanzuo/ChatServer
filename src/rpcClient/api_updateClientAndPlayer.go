package rpcClient

import (
	"fmt"
	_ "github.com/Jordanzuo/ChatServer/src/rpcServer"
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/goutil/logUtil"
)

func updateClientAndPlayer() {
	//参数赋值
	params := make([]interface{}, 2, 2)
	params[0] = getClientCount()
	params[1] = getPlayerCount()

	// 记录日志，以便于排查问题
	logUtil.Log(fmt.Sprintf("发送心跳包,clientCount:%d, playerCount:%d", getClientCount(), getPlayerCount()), logUtil.Debug, true)

	//发送请求
	request(transferObject.UpdateClientAndPlayerCount, params, nil)
}
