package Voucher

import (
	"context"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"
	"xzdp/pkg/response"

	"github.com/bytedance/sonic"
)

func AddDinaryVoucherToDB(v model.TbVoucher) (uint64, error) {
	VourcherQuery := query.TbVoucher
	err := VourcherQuery.Create(&v)
	if err != nil {
		return 0, response.WrapBusinessError(response.ErrDatabase, err, "")
	}
	return v.ID, nil
}

func getVouchersFromDB(shopId int64) ([]*model.TbVoucher, error) {
	VoucherQuery := query.TbVoucher
	dbres, err := VoucherQuery.Where(VoucherQuery.ShopID.Eq(uint64(shopId))).Find()
	return dbres, err
}

func getVouchersFromCache(CacheKey string) ([]*model.TbVoucher, error) {
	CacheRes, err := db.RedisDb.Get(context.Background(), CacheKey).Result()
	var res []*model.TbVoucher
	err = sonic.Unmarshal([]byte(CacheRes), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func setVouchersToCache(CacheKey string, vouchers []*model.TbVoucher) error {
	vouchersJson, err := sonic.Marshal(vouchers)
	if vouchersJson == nil || err != nil {
		return err
	}
	return db.RedisDb.Set(context.Background(), CacheKey, string(vouchersJson), voucherTTL).Err()
}
