package Order

import (
	"context"
	"strconv"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"

	"github.com/bytedance/sonic"
	"gorm.io/gorm"
)

// 根据秒杀优惠券id获取秒杀优惠券信息
func getSeckillVoucherById(tx *query.Query, id int64) (*model.TbSeckillVoucher, error) {
	return tx.TbSeckillVoucher.Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(id))).First()
}

// 往秒杀记录表插入数据
func SeckillVoucherAdd(tx *query.Query, s model.TbVoucherOrder) error {
	return tx.TbVoucherOrder.Create(&s)
}

// 修改秒杀优惠券表
func UpdateSeckillVoucher(tx *query.Query, voucher *model.TbSeckillVoucher) (int64, error) {
	voucher_id, err := strconv.ParseInt(strconv.FormatUint(voucher.VoucherID, 10), 10, 64)
	result, err := tx.TbSeckillVoucher.
		Where(
			tx.TbSeckillVoucher.VoucherID.Eq(uint64(voucher_id)),
			tx.TbSeckillVoucher.Stock.Eq(voucher.Stock), // 条件：voucher_id = ?
			tx.TbSeckillVoucher.Stock.Gt(0),             // 条件：stock > 0（防超卖）
		).
		UpdateColumn(tx.TbSeckillVoucher.Stock, gorm.Expr("stock - 1")) // 库存减 1
	return result.RowsAffected, err
}

// Redis初始化库存
func SetSeckillStockToCache(CacheKey string, stock int) error {
	_, err := db.RedisDb.Set(context.Background(), CacheKey, stock, SeckillVoucherTTL).Result()
	return err
}

// 从redis获取库存
func GetStockfromCache(CacheKey string) int {
	res, err := db.RedisDb.Get(context.Background(), CacheKey).Result()
	var stock int
	if err != nil {
		return -1
	}
	sonic.Unmarshal([]byte(res), stock)
	return stock
}
