package Order

import (
	"context"
	"time"
	"xzdp/db"
)

// 这里我们配合Redis自增和自定义的方式实现
// Redis可以保证id全局唯一；为了增加ID的安全性
// 我们不直接使用Redis自增的数值，而是拼接一些其它信息（时间戳+序列号）。
const (
	coutBits = 32
	incr     = "incr:"
)

// 生成分布式订单ID
func generateOrderId(KeyPrefix string) int64 {
	now := time.Now().Unix()
	// 生产序列号
	// Go语言的时间格式是通过一个特定的参考时间来定义的
	// 这个参考时间是Mon Jan 2 15:04:05 MST 2006
	date := time.Now().Format("2006:01:01")
	count, err := db.RedisDb.Incr(context.Background(), incr+KeyPrefix+":"+date).Result()
	if err != nil {
		return -1
	}
	return now<<coutBits | count
}
