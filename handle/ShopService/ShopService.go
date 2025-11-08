package ShopService

import (
	"errors"
	"log/slog"
	"strconv"
	"time"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	shopKeyPrefix    = "cache:shop"
	shopTypeKey      = ":shopType"
	HotBlogKey       = ":hotBlog"
	shopCacheTTL     = 3 * time.Minute
	shopTypeCacheTTL = 6 * time.Minute
	HotBlogTTL       = 5 * time.Minute
)

func QueryShopById(c *gin.Context) {
	//1.参数验证
	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, response.ErrValidation, "id不能为空！")
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil || idInt < 0 {
		response.Error(c, response.ErrValidation, "无效的商户id")
	}
	//2.从缓存中查找
	shop, err := getShopFromCache(idStr)
	//3.缓存未命中或者出错，从数据库中查找
	if shop == nil || err != nil {
		slog.Error("根据商户id查询cache未命中，开始从数据库中查找")
	}
	shop, err = getShopByIDFromDB(idInt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, response.ErrNotFound, "商户表中查询结果为空")
		return
	}
	//4.将数据写入缓存（可异步）并返回结果
	response.Success(c, shop)
}

func QueryShopTypeList(context *gin.Context) {
	//1.从缓存中查找
	types, err := getShopTypeListFromCache()
	//2.缓存命中，直接返回
	if types != nil && err == nil {
		response.Success(context, types)
		return
	}

	//3.缓存未命中或者出错，从数据库中查找
	slog.Error("商户类型缓存未命中，从数据库查询")
	shopTypeList, err := getShopTypeListFromDB()
	if shopTypeList == nil || err != nil {
		slog.Error("商户类型数据不存在", "err", err)
		response.Error(context, response.ErrDatabase, "查询商户类型失败")
		return
	}

	//4.将数据写入缓存（可异步）并返回结果
	setShopTypeListToCache(shopTypeList)
	response.Success(context, shopTypeList)
}

func GetHotBlog(c *gin.Context) {
	//1.从上下文中拿到页码（路径参数用Query方法获取）
	pageNum := c.Query("current")
	//2.带着页码去缓存查询
	cacheRes, err := getBlogByPageNumFromCache(pageNum)
	if cacheRes != nil && err == nil {
		response.Success(c, cacheRes)
		return
	}
	//3.缓存未命中，从数据库查询
	dbRes, err := getBlogByPageNumFromDB(pageNum)
	if err != nil {
		response.Error(c, response.ErrDatabase, "查询db失败")
		return
	}
	err = setHotBlogToCache(dbRes)
	//5.返回结果
	response.Success(c, dbRes)
}
