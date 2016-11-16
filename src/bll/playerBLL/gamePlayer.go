package playerBLL

import (
	"encoding/json"
	"fmt"
	"github.com/Jordanzuo/ChatServer/src/bll/configBLL"
	"github.com/Jordanzuo/ManageCenterModel_Go/serverGroup"
	"github.com/Jordanzuo/goutil/logUtil"
	"github.com/Jordanzuo/goutil/webUtil"
)

// 获取游戏玩家名称
// serverGroupObj：服务器组对象
// id：玩家Id
// 返回值：
// 玩家名称
// 玩家公会Id
// 是否存在玩家
// 是否能向全区服发送消息
// 错误对象
func GetGamePlayer(serverGroupObj *serverGroup.ServerGroup, id string) (name string, unionId string, isCrossServer, exists bool, err error) {
	// 获取数据库配置
	configObj := configBLL.GetConfig()

	// 定义请求参数
	postDict := make(map[string]string)
	postDict["PlayerId"] = id

	// 连接服务器，以获取数据
	url := fmt.Sprintf("%s/%s", serverGroupObj.Url, configObj.GetPlayerInfoAPI())
	returnBytes, err := webUtil.PostWebData(url, postDict, nil)
	if err != nil {
		logUtil.Log(fmt.Sprintf("验证玩家信息出错，url=%s, 错误信息为：%s", url, err), logUtil.Error, true)
		return
	}

	// 解析返回值
	returnMap := make(map[string]interface{})
	if err = json.Unmarshal(returnBytes, &returnMap); err != nil {
		logUtil.Log(fmt.Sprintf("验证玩家信息出错，反序列化返回值出错，url=%s, return=%s, 错误信息为：%s", url, string(returnBytes), err), logUtil.Error, true)
		return
	}

	// 判断Status状态
	if status_float64, ok := returnMap["Status"].(float64); !ok || int(status_float64) != 0 {
		logUtil.Log(fmt.Sprintf("验证玩家信息出错，返回状态不正确，url=%s, return=%s, 状态信息为：%v", url, string(returnBytes), returnMap["Message"]), logUtil.Error, true)
		return
	}

	// 解析数据
	if valueMap, ok := returnMap["Value"].(map[string]interface{}); ok {
		if name, ok = valueMap["Name"].(string); !ok {
			logUtil.Log(fmt.Sprintf("验证玩家信息出错，没有找到Name属性。url=%s, return=%s", url, string(returnBytes)), logUtil.Error, true)
			err = fmt.Errorf("验证玩家信息出错，没有找到Name属性")
			return
		}

		if unionId, ok = valueMap["UnionId"].(string); !ok {
			logUtil.Log(fmt.Sprintf("验证玩家信息出错，没有找到UnionId属性。url=%s, return=%s", url, string(returnBytes)), logUtil.Error, true)
			err = fmt.Errorf("验证玩家信息出错，没有找到UnionId属性")
			return
		}

		if isCrossServer, ok = valueMap["IsCrossServer"].(bool); !ok {
			logUtil.Log(fmt.Sprintf("验证玩家信息出错，没有找到IsCrossServer属性。url=%s, return=%s", url, string(returnBytes)), logUtil.Error, true)
			err = fmt.Errorf("验证玩家信息出错，没有找到IsCrossServer属性")
			return
		}

		exists = true
	}

	return
}
