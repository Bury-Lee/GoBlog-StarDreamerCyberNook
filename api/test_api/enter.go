package test_api

import (
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
