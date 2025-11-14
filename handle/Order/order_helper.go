package Order

import (
	"strconv"
	"xzdp/dal/model"
	"xzdp/dal/query"

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
func UpdateSeckillVoucher(tx *query.Query, voucherID string) (int64, error) {
	// sqlStmt := "UPDATE tb_seckill_voucher t SET stock = stock - 1 WHERE t.voucher_id = ?"
	// tx.TbSeckillVoucher.Exec(sqlStmt, voucherID)
	voucher_id, err := strconv.ParseInt(voucherID, 10, 64)
	result, err := tx.TbSeckillVoucher.
		Where(tx.TbSeckillVoucher.VoucherID.Eq(uint64(voucher_id))).    // 条件：voucher_id = ?
		Where(tx.TbSeckillVoucher.Stock.Gt(0)).                         // 条件：stock > 0（防超卖）
		UpdateColumn(tx.TbSeckillVoucher.Stock, gorm.Expr("stock - 1")) // 库存减 1
	return result.RowsAffected, err
}
