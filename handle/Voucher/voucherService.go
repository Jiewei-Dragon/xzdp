package Voucher

import (
	"log/slog"
	"strconv"
	"time"
	"xzdp/dal/model"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	voucherKeyPrefix = "voucher:shop:"
	voucherTTL       = 30 * time.Second
)

func AddVoucher(c *gin.Context) {
	var voucherReq model.TbVoucher
	err := c.ShouldBind(voucherReq)
	VoucherType := voucherReq.Type
	var id uint64
	if VoucherType == 0 {
		id, err = AddDinaryVoucherToDB(voucherReq)
	} // } else if VoucherType == 1 {

	// }
	response.HandleBusinessResult(c, err, gin.H{"voucherId": id})
}

func GetVouchersByShopId(c *gin.Context) {
	shopId := c.Param("shopId")
	if shopId == "" {
		response.Error(c, response.ErrValidation, "无效参数")
	}
	CacheKey := voucherKeyPrefix + shopId
	CacheRes, err := getVouchersFromCache(CacheKey)
	if CacheRes != nil && err == nil {
		response.Success(c, CacheRes)
		return
	}
	Id, err := strconv.ParseInt(shopId, 10, 64)
	if err != nil {
		return
	}
	resDb, err := getVouchersFromDB(Id)
	if err != nil {
		response.Error(c, response.ErrDatabaseNotFind)
	}
	var res []VoucherDTO
	for _, v := range resDb {
		res = append(res, VouchermodelToDTO(v))
	}
	response.Success(c, res)
	err = setVouchersToCache(CacheKey, resDb)
	if err != nil {
		slog.Log(c, 1, "博客缓存失败")
	}
}
