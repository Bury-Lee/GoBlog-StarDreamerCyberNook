package test_api

import (
	"StarDreamerCyberNook/common/response"
	"StarDreamerCyberNook/global"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

type TestApi struct {
}

func (TestApi) TestView(c *gin.Context) {
	// response.FailWithMsg("测试失败", c)
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "测试成功",
	})
}

// GetConfig 查看运行时配置(测试用):ES/AI 开关与服务地址表。
func (TestApi) GetConfig(c *gin.Context) {
	response.OkWithData(gin.H{
		"esEnabled": global.Config.ES.Enabled,
		"aiEnabled": global.Config.AI.Enable,
		"services":  global.Config.Services,
	}, c)
}

// SetConfig 直接修改运行时配置(测试用):请求体 JSON 合并进 global.Config(仅覆盖出现的字段)。
// 例:{"es":{"enabled":true}} 或 {"services":{"search":["127.0.0.1:9220"]}}
func (TestApi) SetConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.FailWithMsg("读取请求体失败", c)
		return
	}
	if err := json.Unmarshal(body, global.Config); err != nil {
		response.FailWithMsg("配置解析失败: "+err.Error(), c)
		return
	}
	response.OkWithMsg("配置已更新", c)
}
