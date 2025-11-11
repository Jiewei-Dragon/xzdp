package Voucher

import "xzdp/dal/model"

// type VoucherDTO struct {
// 	id           int
// 	shop_id      int	'json:"shopId"'
// 	title        string
// 	sub_title    string
// 	rules        string
// 	pay_value    int
// 	actual_value int
// 	Type         int
// 	status       int
// 	create_time  string
// 	update_time  string
// }

type VoucherDTO struct {
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
