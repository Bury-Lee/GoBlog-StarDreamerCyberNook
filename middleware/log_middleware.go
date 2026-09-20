// middleware/log_middleware.go
package middleware

import (
	"StarDreamerCyberNook/service/log_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseWriter struct {
	gin.ResponseWriter
	Body []byte
}

// maxLogBodySize 日志中最多捕获的响应体大小,避免大响应(如图片)占用内存
const maxLogBodySize = 4 * 1024

func (w *ResponseWriter) Write(data []byte) (int, error) { //重写Write方法以捕获响应体
	if len(w.Body) < maxLogBodySize {
		remaining := maxLogBodySize - len(w.Body)
		if len(data) > remaining {
			w.Body = append(w.Body, data[:remaining]...)
		} else {
			w.Body = append(w.Body, data...)
		}
	}
	return w.ResponseWriter.Write(data)
}

func (w *ResponseWriter) Header() http.Header {
	//必须返回底层真实 Writer 的 Header,否则 Content-Type 等响应头会被丢弃
	return w.ResponseWriter.Header()
}

func LogMiddleware(c *gin.Context) {
	log := log_service.NewActionLog(c) // 创建日志实例
	log.SetRequest(c)

	c.Set("log", log)

	// 3. 替换 c.Writer 为我们的自定义响应写入器
	res := &ResponseWriter{
		ResponseWriter: c.Writer,
	}
	c.Writer = res
	// 4. 继续执行下一个中间件或路由处理函数
	c.Next()
	// 5. 处理响应后：打印响应体
	log.SetResponse(res.Body)
	log.SetResponseHeader(res.ResponseWriter.Header())
	log.MiddlewareSave()

}
