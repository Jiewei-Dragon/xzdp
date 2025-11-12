package Voucher

import (
	"context"
	"time"
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

func AddSeckillVoucherToDB(v VoucherDTO) (uint64, error) {

	start, err := time.Parse(timeLayout, v.BeginTime)
	if err != nil {
		return 0, response.WrapBusinessError(response.ErrValidation, err, "BeginTime "+timeFormatError)
	}
	end, err := time.Parse(timeLayout, v.EndTime)
	if err != nil {
		return 0, response.WrapBusinessError(response.ErrValidation, err, "EndTime "+timeFormatError)
	}
	// 验证时间逻辑
	if !end.After(start) {
		return 0, response.WrapBusinessError(response.ErrValidation, nil, "结束时间必须晚于开始时间！")
	}

	voucherDbModel := DTOToVoucherModel(v)
	// 这里要用到事务确保都成功或者失败
	q := query.Use(db.DBEngine) //初始化查询对象（Query）并关联数据库连接
	err = q.Transaction(func(tx *query.Query) error {
		// 1.先insert到Voucher表
		err := tx.TbVoucher.Create(&voucherDbModel)
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "Voucher表插入失败")
		}
		// 2.再insert到SeckillVoucher表
		//2.再添加信息到秒杀卷表 tb_seckill_voucher
		seckillDbModel := model.TbSeckillVoucher{
			VoucherID: voucherDbModel.ID,
			Stock:     int32(v.Stock),
			BeginTime: start,
			EndTime:   end,
		}
		err = tx.TbSeckillVoucher.Create(&seckillDbModel)
		if err != nil {
			return response.WrapBusinessError(response.ErrDatabase, err, "SeckillVoucher表插入失败")
		}
		return nil
	})

	return voucherDbModel.ID, nil
}

func getVouchersFromDB(shopId int64) ([]*VoucherDTO, error) {
	v := query.TbVoucher
	sv := query.TbSeckillVoucher
	var result []*VoucherDTO
	err := v.LeftJoin(sv, v.ID.EqCol(sv.VoucherID)).
		Where(v.ShopID.Eq(uint64(shopId))).
		Select(
			v.ID, v.ShopID, v.Title, v.SubTitle, v.Rules,
			v.PayValue, v.ActualValue, v.Type, v.Status,
			v.CreateTime, v.UpdateTime,
			sv.Stock, sv.BeginTime, sv.EndTime, // 秒杀相关字段
		).
		Scan(&result)
	return result, err
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

func setVouchersToCache(CacheKey string, vouchers []*VoucherDTO) error {
	vouchersJson, err := sonic.Marshal(vouchers)
	if vouchersJson == nil || err != nil {
		return err
	}
	return db.RedisDb.Set(context.Background(), CacheKey, string(vouchersJson), voucherTTL).Err()
}
