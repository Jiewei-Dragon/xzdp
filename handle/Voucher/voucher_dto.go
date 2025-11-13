package Voucher

import (
	"xzdp/dal/model"
)

// 定义接收联合查询结果的结构体
// type SeckillVoucherVO struct {
// 	ShopID      int       `json:"shopId"`      // 对应 v.shop_id
// 	Title       string    `json:"title"`       // 对应 v.title
// 	SubTitle    string    `json:"subTitle"`    // 对应 v.sub_title
// 	Rules       string    `json:"rules"`       // 对应 v.rules
// 	PayValue    int       `json:"payValue"`    // 对应 v.pay_value
// 	ActualValue int       `json:"actualValue"` // 对应 v.actual_value
// 	Type        int       `json:"type"`        // 对应 v.type
// 	Stock       int       `json:"stock"`       // 对应 s.stock
// 	BeginTime   time.Time `json:"beginTime"`   // 对应 s.begin_time
// 	EndTime     time.Time `json:"endTime"`     // 对应 s.end_time
// }

type VoucherDTO struct {
	ID          int    `json:"id"`     //优惠券ID
	ShopId      int    `json:"shopId"` //关联的商店id
	Title       string `json:"title"`
	SubTitle    string `json:"subTitle"`
	Rules       string `json:"rules"`
	PayValue    int    `json:"payValue"` //优惠的价格
	ActualValue int    `json:"actualValue"`
	Type        int    `json:"type"`  //优惠卷类型
	Stock       int    `json:"stock"` //库存
	BeginTime   string `json:"beginTime"`
	EndTime     string `json:"endTime"`
}

func VouchermodelToDTO(m *model.TbVoucher) VoucherDTO {
	return VoucherDTO{
		ShopId:      int(m.ShopID),
		Title:       m.Title,
		SubTitle:    m.SubTitle,
		Rules:       m.Rules,
		PayValue:    int(m.PayValue), //优惠的价格
		ActualValue: int(m.ActualValue),
		Type:        int(m.Type),
	}
}

func DTOToVoucherModel(v VoucherDTO) model.TbVoucher {
	return model.TbVoucher{
		ShopID:      uint64(v.ShopId),
		Title:       v.Title,
		SubTitle:    v.SubTitle,
		Rules:       v.Rules,
		PayValue:    uint64(v.PayValue), //优惠的价格
		ActualValue: int64(v.ActualValue),
		Type:        uint32(v.Type),
	}
}

func VoucherModelToDTO(v model.TbVoucher) VoucherDTO {
	return VoucherDTO{
		ShopId:      int(v.ShopID),
		Title:       v.Title,
		SubTitle:    v.SubTitle,
		Rules:       v.Rules,
		PayValue:    int(v.PayValue),
		ActualValue: int(v.ActualValue),
		Type:        int(v.Type),
	}
}
