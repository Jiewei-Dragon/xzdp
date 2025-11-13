package Order

import (
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"
)

// 根据秒杀优惠券id获取秒杀优惠券信息
func getSeckillVoucherById(id int64) (*model.TbSeckillVoucher, error) {
	queryVoucher := query.TbSeckillVoucher
	return queryVoucher.Where(queryVoucher.VoucherID.Eq(uint64(id))).First()
}

// 往秒杀记录表插入数据
func SeckillVoucherAdd(s model.TbVoucherOrder) error {
	queryVoucherOrder := query.TbVoucherOrder
	return queryVoucherOrder.Create(&s)
}

// 修改秒杀优惠券表
func UpdateSeckillVoucher(voucherID string) error {
	sqlStmt := "UPDATE tb_seckill_voucher t SET stock = stock - 1 WHERE t.voucher_id = ?"
	return db.DBEngine.Exec(sqlStmt, voucherID).Error
}
