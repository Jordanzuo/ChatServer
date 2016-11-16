package rpcClient

import (
	"github.com/Jordanzuo/ChatServerModel/src/transferObject"
	"github.com/Jordanzuo/goutil/debugUtil"
)

func login() {
	params := make([]interface{}, 1, 1)
	params[0] = chatServerPublicAddress

	//发送Login消息
	debugUtil.Println("\nSend Login")

	request(transferObject.Login, params, loginCallback)
}

func loginCallback(data interface{}) {
	debugUtil.Println("Login success")
	loginSucceedCh <- 1
}
