package Voucher

import (
	"log/slog"
	"strconv"
	"time"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	timeLayout       = "2006-01-02 15:04:05"
	timeFormatError  = "time format error, must be like 2006-01-02 15:04:05"
	voucherKeyPrefix = "voucher:shop:"
	voucherTTL       = 30 * time.Second
)

func AddVoucher(c *gin.Context) {
	var voucherReq VoucherDTO
	err := c.BindJSON(&voucherReq)
	VoucherType := voucherReq.Type
	var id uint64
	if VoucherType == 0 {
		id, err = AddDinaryVoucherToDB(DTOToVoucherModel(voucherReq))
	} else if VoucherType == 1 {
		id, err = AddSeckillVoucherToDB(voucherReq)
	}
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
	response.Success(c, resDb)
	err = setVouchersToCache(CacheKey, resDb)
	if err != nil {
		slog.Log(c, 1, "博客缓存失败")
	}
}
