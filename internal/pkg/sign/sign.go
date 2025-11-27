package sign

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// VerifyNotifySign 校验支付回调签名
// params: 请求参数的 map（不包含 sign 字段）
// secret: 签名密钥
// sign: 待校验的签名
func VerifyNotifySign(params map[string]string, secret, sign string) bool {
	if sign == "" {
		return false
	}
	calculatedSign := GenerateSign(params, secret)
	return calculatedSign == sign
}

// GenerateSign 生成签名
// 1. 将参数按 key 排序
// 2. 拼接成 key1=value1&key2=value2&key=secret 格式
// 3. MD5 加密并转大写
func GenerateSign(params map[string]string, secret string) string {
	// 获取所有 key 并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" || params[k] == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}
	builder.WriteString("&key=")
	builder.WriteString(secret)

	// MD5 加密
	signStr := builder.String()
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// ParamsToMap 将结构体转为 map（用于签名）
func ParamsToMap(data interface{}) map[string]string {
	// 这里简化处理，实际项目中可以用反射或手动构造
	// 为了简单，在调用时手动构造 map
	return nil
}

// Example 使用示例
func Example() {
	params := map[string]string{
		"order_id":     "123456",
		"out_trade_no": "wx_trade_001",
		"user_id":      "1001",
		"quantity":     "1",
		"status":       "paid",
		"paid_at":      "2025-11-27T17:00:00Z",
	}
	secret := "your-secret-key"

	sign := GenerateSign(params, secret)
	fmt.Println("Generated Sign:", sign)

	isValid := VerifyNotifySign(params, secret, sign)
	fmt.Println("Sign Valid:", isValid)
}
