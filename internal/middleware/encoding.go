package middleware

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// FixEncoding 中间件：检测非UTF-8请求体并从GBK转换为UTF-8
func FixEncoding() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}

		// 只处理 JSON 请求
		ct := c.GetHeader("Content-Type")
		if !strings.Contains(ct, "json") {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil || len(body) == 0 {
			c.Next()
			return
		}

		// 如果已经是合法UTF-8，不处理
		if utf8.Valid(body) {
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Next()
			return
		}

		// 尝试从GBK转换为UTF-8
		reader := transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder())
		converted, err := io.ReadAll(reader)
		if err != nil {
			// 转换失败，保留原始内容
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(converted))
		c.Next()
	}
}
