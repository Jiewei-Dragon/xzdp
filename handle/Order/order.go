package Order

import (
	"context"
	"errors"
	"strconv"
	"time"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"
	"xzdp/middleware"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
)

// 这里我们配合Redis自增和自定义的方式实现
// Redis可以保证id全局唯一；为了增加ID的安全性
// 我们不直接使用Redis自增的数值，而是拼接一些其它信息（时间戳+序列号）。
const (
	coutBits = 32
	incr     = "incr:"
)

type seckillRequest struct {
	UserId    int `json:"userId"`
	VoucherId int `json:"voucherId"`
}

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

// 秒杀优惠券
func SeckillVouchers(c *gin.Context) {
	voucherIdStr := c.Param("id")
	voucherIdInt, err := strconv.Atoi(voucherIdStr)
	if err != nil {
		response.HandleBusinessError(c, err)
	}
	userId := c.GetInt64(middleware.CtxKeyUserId)
	// 1.先判断是否已经过期
	reqTime := time.Now()
	voucherOld, err := getSeckillVoucherById(int64(voucherIdInt))
	if reqTime.After(voucherOld.EndTime) || reqTime.Before(voucherOld.BeginTime) {
		response.Error(c, response.ErrValidation, "不再秒杀优惠券时间范围内")
		return
	}
	// 2.再判断库存
	if voucherOld.Stock <= 0 {
		response.Error(c, response.ErrValidation, "优惠券已经没啦，下次再快一点")
		return
	}
	// 3.秒杀优惠券涉及到两张表，用事务确保原子性
	q := query.Use(db.DBEngine)
	globalId := generateOrderId("order")
	err = q.Transaction(func(tx *query.Query) error {
		// 3.0通过CAS乐观锁修改库存
		voucherNew, err := getSeckillVoucherById(int64(voucherIdInt))
		if voucherOld.Stock != voucherNew.Stock {
			return errors.New("网络开小差啦，请重试")
		}
		// 3.1 先减少库存
		err = UpdateSeckillVoucher(voucherIdStr)
		if err != nil {
			return err
		}
		// 3.2 新增秒杀记录
		sv := model.TbVoucherOrder{
			ID:         globalId,
			UserID:     uint64(userId),
			VoucherID:  uint64(voucherIdInt),
			PayType:    1,
			Status:     0,
			CreateTime: reqTime,
			PayTime:    reqTime,
			UseTime:    reqTime,
			RefundTime: reqTime,
			UpdateTime: reqTime,
		}
		err = SeckillVoucherAdd(sv)
		return err
	})
	if err != nil {
		response.Error(c, response.ErrDatabase, err.Error())
		return
	}
	response.Success(c, gin.H{"orderId": strconv.FormatInt(globalId, 10)})
}
