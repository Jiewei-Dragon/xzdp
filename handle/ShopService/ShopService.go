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
	shopTypeCacheTTL = 0 * time.Minute
	HotBlogTTL       = 5 * time.Minute
	ShopPageSize     = 5
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
	CacheKey := shopKeyPrefix + ":Id:" + idStr
	//2.从缓存中查找
	shop, err := getShopByIdFromCache(CacheKey)
	//3.缓存未命中或者出错，从数据库中查找
	if shop == nil || err != nil {
		slog.Error("根据商户id查询cache未命中，开始从数据库中查找")
	}
	shop, err = getShopByIdFromDB(idInt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, response.ErrNotFound, "商户表中查询结果为空")
		return
	}
	//4.将数据写入缓存（可异步）并返回结果
	err = setShopByIdtoCache(CacheKey, *shop)
	if err != nil {
		return
	}
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
	shopTypeList, err := getShopTypeListFromDB()
	if shopTypeList == nil || err != nil {
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
	if err != nil {
		//5.返回结果
		response.Success(c, dbRes)
	} else {
		response.Error(c, response.ErrDatabaseNotFind, "已经没有更多了")
	}

}

func GetShopByTypeId(c *gin.Context) {
	//1.解析参数并验证
	typeId, current, x, y, sortBy := c.Query("typeId"), c.Query("current"), c.Query("x"), c.Query("y"), c.Query("sortBy")
	if typeId == "" || current == "" || x == "" || y == "" {
		response.Error(c, response.ErrValidation, "参数不能为空")
		return
	}
	typeIdInt, err := strconv.Atoi(typeId)
	if err != nil {
		response.Error(c, response.ErrValidation, "typeId必须为数字")
		return
	}
	currentInt, err := strconv.Atoi(current)
	if err != nil {
		response.Error(c, response.ErrValidation, "currentId必须为数字")
		return
	}
	if sortBy == "" {
		sortBy = "empty"
	}
	// 3.从缓存中查找
	CacheKey := shopKeyPrefix + ":typeId:" + typeId + ":sortBy:" + sortBy + ":current:" + current
	cacheRes, err := getShopsByTypeIdFromCache(CacheKey)
	if cacheRes != nil && err == nil {
		response.Success(c, cacheRes)
		return
	}

	// 4.缓存未命中，从数据库查询
	dbRes, err := getShopsByTypeIdFromDB(typeIdInt, sortBy, currentInt)
	if err != nil {
		slog.Error("数据库查询失败", "typeIdInt", typeIdInt, "err", err)
		response.Error(c, response.ErrDatabaseNotFind)
		return
	}

	// 5.只有当查询结果不为空时才写入缓存，避免缓存空数组导致重复渲染
	if len(dbRes) > 0 {
		err = setShopsByTypeIdToCache(CacheKey, dbRes)
		if err != nil {
			slog.Error("写入缓存失败", "CacheKey", CacheKey, "err", err)
			// 缓存失败不影响返回结果，继续执行
		}
		response.Success(c, dbRes)
	} else {
		response.Error(c, response.ErrDatabaseNotFind, "已经没有更多了")
	}
}
