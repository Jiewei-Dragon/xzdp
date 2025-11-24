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

// 压测时候用
type reqBody struct {
	UserId int64 `json:"userId" binding:"omitempty"`
}

const (
	SeckillVoucherKeyPrefix = "SeckillVoucher:"
	SeckillVoucherTTL       = 30 * time.Minute
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
	seckill := query.TbSeckillVoucher
	CacheKey := SeckillVoucherKeyPrefix + voucherIdStr
	err = c.BindJSON(&reqbody)
	if err != nil {
		response.HandleBusinessError(c, err)
	}
	userId := reqbody.UserId
	//------------------------------------------------
	// 先判断是否已经拥有了优惠券
	order := query.TbVoucherOrder
	res, err := order.Where(order.UserID.Eq(uint64(userId))).Find()
	if len(res) > 0 {
		response.Error(c, response.ErrValidation, "每个用户限购一张该优惠券")
		return
	}
	// 1.判断是否已经过期
	reqTime := time.Now()
	//这里就不能用helper里面的方法了，因为里面的方法需要在事务下进行
	CacheStock := GetStockfromCache(CacheKey)
	if CacheStock < 0 {
		seckill, err := seckill.Where(seckill.VoucherID.Eq(uint64(voucherIdInt))).First()
		if seckill == nil || err != nil {
			return
		}
		SetSeckillStockToCache(CacheKey, int(seckill.Stock))
		// 1.再判断库存
		if reqTime.After(seckill.EndTime) || reqTime.Before(seckill.BeginTime) {
			response.Error(c, response.ErrValidation, "不在秒杀优惠券时间范围内")
			return
		}
		// 2.再判断库存
		if seckill.Stock <= 0 {
			response.Error(c, response.ErrValidation, "优惠券已经没啦，下次再快一点")
			return
		}
	}
	// DECR原子减1，返回减后的值（避免并发问题）
	remainStock, err := db.RedisDb.Decr(context.Background(), SeckillVoucherKeyPrefix+voucherIdStr).Result()
	if err != nil {
		response.Error(c, response.ErrValidation, "网络繁忙，请重试")
		return
	}
	if remainStock < 0 {
		response.Error(c, response.ErrValidation, "优惠券已经没啦，下次再快一点")
		return
	}
	// 3. 预扣减成功后，再执行数据库事务（这一步才走到数据库）
	q := query.Use(db.DBEngine)
	globalId := generateOrderId("order")
	UserLockMap.Lock(int(userId))
	err = q.Transaction(func(tx *query.Query) error {
		voucher, err := getSeckillVoucherById(tx, int64(voucherIdInt))
		// 3.1 通过CAS乐观锁修改库存
		RowsAffected, err := UpdateSeckillVoucher(tx, voucher)
		// SQL 更新返回的 RowsAffected = 0说明 库存数没同步或者库存数被其他线程扣到0了
		if RowsAffected == 0 || err != nil {
			response.Error(c, response.ErrValidation, "你的优惠券被其他人抢走啦，请重试！")
			// 回滚
			db.RedisDb.Incr(context.Background(), CacheKey)
			c.Abort() // 终止请求，不再执行后续代码
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
	UserLockMap.Unlock(int(userId))
	if err != nil {
		if !c.IsAborted() {
			//回滚Redis里的库存
			db.RedisDb.Incr(context.Background(), CacheKey)
			response.Error(c, response.ErrDatabase, "秒杀事务异常")
			c.Abort()
		}
	}
	if !c.IsAborted() {
		response.Success(c, gin.H{"orderId": strconv.FormatInt(globalId, 10)})
	}
}
