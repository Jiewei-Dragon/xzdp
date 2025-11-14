package Order

import (
	"context"
	"strconv"
	"time"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"
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

type reqBody struct {
	UserId int64 `json:"userId" binding:"omitempty"`
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
	// userId := c.GetInt64(middleware.CtxKeyUserId)

	//-----------------------------------------------
	var reqbody reqBody
	err = c.BindJSON(&reqbody)
	if err != nil {
		response.HandleBusinessError(c, err)
	}
	userId := reqbody.UserId
	//-----------------------------------------------

	// 0.先判断是否已经拥有了优惠券
	order := query.TbVoucherOrder
	res, err := order.Where(order.UserID.Eq(uint64(userId))).Find()
	if len(res) > 0 {
		response.Error(c, response.ErrValidation, "每个用户限购一张该优惠券")
		return
	}
	// 1.判断是否已经过期
	reqTime := time.Now()
	//这里就不能用helper里面的方法了，因为里面的方法需要在事务下进行
	seckill := query.TbSeckillVoucher
	voucherOld, err := seckill.Where(seckill.VoucherID.Eq(uint64(voucherIdInt))).First()
	if reqTime.After(voucherOld.EndTime) || reqTime.Before(voucherOld.BeginTime) {
		response.Error(c, response.ErrValidation, "不在秒杀优惠券时间范围内")
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
		voucherNew, err := getSeckillVoucherById(tx, int64(voucherIdInt))
		if voucherOld.Stock != voucherNew.Stock {
			response.Error(c, response.ErrValidation, "网络开小差啦，请重试")
			return err
		}
		// 3.1 先减少库存
		RowsAffected, err := UpdateSeckillVoucher(tx, voucherIdStr)
		if RowsAffected < 0 || err != nil {
			response.Error(c, response.ErrValidation, "你来晚了，优惠券已经没了！")
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
		SeckillVoucherAdd(tx, sv)
		return err
	})
	if err != nil {
		response.Error(c, response.ErrDatabase, "秒杀事务异常")
		return
	}
	response.Success(c, gin.H{"orderId": strconv.FormatInt(globalId, 10)})
}
