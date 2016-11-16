package rpcServer

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Jordanzuo/ChatServerModel/src/channelType"
	"github.com/Jordanzuo/ChatServerModel/src/commandType"
	"github.com/Jordanzuo/ChatServerModel/src/player"
	"github.com/Jordanzuo/ChatServerModel/src/serverResponseObject"
	"github.com/Jordanzuo/goutil/logUtil"
)

// 处理需要客户端发送的数据
// clientObj：客户端对象
func handleSendData(clientObj *Client) {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	for {
		//连接是否断开
		if clientObj.getConnStatus() == con_Close {
			break
		}

		//是否是最后一条消息
		connStatus := clientObj.getConnStatus()

		// 是否被处理过
		handled := false

		// 优先处理高优先级的数据，如果发送出现错误，表示连接已经断开，则退出方法；如果没有待处理的数据，则退出循环
		for {
			if responseObject, exists := clientObj.getSendData(); exists {
				handled = true
				if err := clientObj.sendMessage(responseObject); err != nil {
					return
				}
			} else {
				break
			}
		}

		// 当没有高优先级的数据时才处理低优先级的数据，如果发送出现错误，表示连接已经断开，则退出方法；
		if responseObject, exists := clientObj.getSendData_LowPriority(); exists {
			handled = true
			if err := clientObj.sendMessage(responseObject); err != nil {
				return
			}
		}

		// 如果本轮没有被处理过，则休眠5ms
		if !handled {
			time.Sleep(5 * time.Millisecond)
			if connStatus == con_WaitForClose {
				clientObj.setConnStatus(con_Close)
			}
		}
	}
}

// 处理客户端收到的数据
// clientObj：客户端对象
func handleReceiveData(clientObj *Client) {
	for {
		// 获取有效的消息
		message, exists := clientObj.getReceiveData()
		if !exists {
			break
		}

		// 处理数据，如果长度为0则表示心跳包；否则处理请求内容
		if len(message) == 0 {
			clientObj.WriteLog("收到心跳消息")
			continue
		} else {
			handleRequest(clientObj, message)
		}
	}
}

// 处理客户端连接
// conn：客户端连接对象
func handleConn(conn net.Conn) {
	// 处理内部未处理的异常，以免导致主线程退出，从而导致系统崩溃
	defer func() {
		if r := recover(); r != nil {
			logUtil.LogUnknownError(r)
		}
	}()

	// 创建客户端对象
	clientObj := newClient(conn)

	// 将客户端对象添加到客户端增加的channel中
	registerClient(clientObj)

	// 启动处理数据的Goroutine
	go handleSendData(clientObj)

	// 释放client对象
	defer func() {
		disconnectByClient(clientObj)
	}()

	// 无限循环，不断地读取数据，解析数据，处理数据
	for {
		// 先读取数据，每次读取1024个字节
		readBytes := make([]byte, 1024)

		// Read方法会阻塞，所以不用考虑异步的方式
		n, err := conn.Read(readBytes)
		if err != nil {
			if err == io.EOF {
				clientObj.WriteLog(fmt.Sprintf("读取消息时收到断开错误：%s，本次读取的字节数为：%d", err, n))
			} else {
				clientObj.WriteLog(fmt.Sprintf("读取消息错误：%s，本次读取的字节数为：%d", err, n))
			}

			break
		}

		// 将读取到的数据追加到已获得的数据的末尾
		clientObj.appendReceiveData(readBytes[:n])

		// 处理数据
		handleReceiveData(clientObj)
	}
}

// 处理客户端请求
// clientObj：对应的客户端对象
// request：请求内容字节数组(json格式)
// 返回值：无
func handleRequest(clientObj *Client, request []byte) {
	responseObj := serverResponseObject.NewResponseObject(commandType.Login)

	defer func() {
		// 如果不成功，则向客户端发送数据；因为成功已经通过对应的方法发送结果，故不通过此处
		if responseObj.Code != serverResponseObject.Con_Success {
			// 如果是客户端数据错误，则将客户端请求数据记录下来
			if responseObj.Code == serverResponseObject.Con_ClientDataError {
				logUtil.Log(fmt.Sprintf("请求的数据为：%s, 返回的结果为客户端数据错误", string(request)), logUtil.Error, true)
			}

			//调用发送消息接口
			ResponseResult(clientObj, responseObj, Con_HighPriority)
		}
	}()

	// 定义变量
	var requestMap map[string]interface{}
	var commandMap map[string]interface{}
	var playerObj *player.Player
	var id string
	var name string
	var unionId string
	var extraMsg string
	var sign string
	var partnerId int
	var serverId int
	var message string
	var toPlayerId string
	var _commandType commandType.CommandType
	var _channelType channelType.ChannelType
	var exists bool
	var ok bool
	var err error

	// 解析请求字符串
	if err = json.Unmarshal(request, &requestMap); err != nil {
		logUtil.Log(fmt.Sprintf("反序列化出错，错误信息为：%s", err), logUtil.Error, true)
		responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
		return
	}

	// 解析CommandType
	if commandType_float, ok := requestMap["CommandType"].(float64); !ok {
		logUtil.Log("CommandType不是int类型", logUtil.Error, true)
		responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
		return
	} else {
		_commandType = commandType.CommandType(int(commandType_float))
	}

	// 设置responseObject的CommandType
	responseObj.SetCommandType(_commandType)

	// 如果不是Login方法，则判断Client对象所对应的玩家对象是否存在（因为当是Login方法时，Player对象尚不存在）
	if _commandType != commandType.Login {
		if clientObj.GetPlayerId() == "" {
			responseObj.SetResultStatus(serverResponseObject.Con_NoLogin)
			return
		}

		playerObj, exists, err = getPlayer(clientObj.GetPlayerId(), false)
		if err != nil {
			responseObj.SetResultStatus(serverResponseObject.Con_DataError)
			return
		}

		if !exists {
			responseObj.SetResultStatus(serverResponseObject.Con_NoLogin)
			return
		}
	}

	// 解析Command(是map[string]interface{}类型)；只有当不是Logout方法时才解析，因为Logout时Command为空
	if _commandType != commandType.Logout {
		if commandMap, ok = requestMap["Command"].(map[string]interface{}); !ok {
			logUtil.Log("commandMap不是map类型", logUtil.Error, true)
			responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
			return
		}

		// 解析出详细的参数
		if id_interface, exists := commandMap["Id"]; exists {
			if id, ok = id_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("id:%v不是string类型", id_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}

		if name_interface, exists := commandMap["Name"]; exists {
			if name, ok = name_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("name:%v不是string类型", name_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}

		if unionId_interface, exists := commandMap["UnionId"]; exists {
			if unionId, ok = unionId_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("unionId:%v不是string类型", unionId_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}

		if extraMsg_interface, exists := commandMap["ExtraMsg"]; exists {
			if extraMsg, ok = extraMsg_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("extraMsg:%v不是string类型", extraMsg_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}

		if sign_interface, exists := commandMap["Sign"]; exists {
			if sign, ok = sign_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("sign:%v不是string类型", sign_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}

		if partnerId_interface, exists := commandMap["PartnerId"]; exists {
			if partnerId_float64, ok := partnerId_interface.(float64); !ok {
				logUtil.Log(fmt.Sprintf("partnerId:%v不是float64类型", partnerId_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			} else {
				partnerId = int(partnerId_float64)
			}
		}

		if serverId_interface, exists := commandMap["ServerId"]; exists {
			if serverId_float64, ok := serverId_interface.(float64); !ok {
				logUtil.Log(fmt.Sprintf("serverId:%v不是float64类型", serverId_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			} else {
				serverId = int(serverId_float64)
			}
		}

		if channelType_interface, exists := commandMap["ChannelType"]; exists {
			if channelType_float64, ok := channelType_interface.(float64); !ok {
				logUtil.Log(fmt.Sprintf("channelType:%v不是float64类型", channelType_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			} else {
				_channelType = channelType.ChannelType(channelType_float64)
			}
		}

		if message_interface, exists := commandMap["Message"]; exists {
			if message, ok = message_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("message:%v不是string类型", message_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}

		if toPlayerId_interface, exists := commandMap["ToPlayerId"]; exists {
			if toPlayerId, ok = toPlayerId_interface.(string); !ok {
				logUtil.Log(fmt.Sprintf("toPlayerId:%v不是string类型", toPlayerId_interface), logUtil.Error, true)
				responseObj.SetResultStatus(serverResponseObject.Con_ClientDataError)
				return
			}
		}
	}

	// 调用方法
	switch _commandType {
	case commandType.Login:
		responseObj = loginHandler(clientObj, id, name, unionId, extraMsg, sign, partnerId, serverId)
	case commandType.Logout:
		responseObj = logoutHandler(clientObj, playerObj)
	case commandType.UpdatePlayerInfo:
		responseObj = updatePlayerInfoHandler(clientObj, playerObj, name, unionId, extraMsg)
	case commandType.SendMessage:
		responseObj = sendMessageHandler(clientObj, playerObj, _channelType, message, toPlayerId)
	default:
		logUtil.Log("未找到该方法", logUtil.Error, true)
		responseObj.SetResultStatus(serverResponseObject.Con_CommandTypeNotDefined)
	}
}
