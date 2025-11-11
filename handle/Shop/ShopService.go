package Shop

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"
	"xzdp/dal/query"
	"xzdp/db"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	shopKeyPrefix    = "cache:shop"
	shopTypeKey      = ":shopType"
	BlogPrefix       = "blog"
	HotBlogKey       = ":hotBlog"
	shopCacheTTL     = 3 * time.Minute
	shopTypeCacheTTL = 0 * time.Minute
	HotBlogTTL       = 5 * time.Minute
	ShopPageSize     = 10
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
		response.Error(context, response.ErrDatabaseNotFind, "查询商户类型失败")
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
		slog.Log(c, 1, "博客缓存失败")
	}
	response.Success(c, dbRes)

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

// PUT /api/v1/shop
func UpdateShop(c *gin.Context) {
	var shop ShopRequest
	err := c.ShouldBindJSON(&shop)
	if err != nil {
		slog.Error("bindjson bad", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	// 更新时必须提供有效的ID
	if shop.ID == 0 {
		response.Error(c, response.ErrValidation, "shop id is required for update")
		return
	}
	update(c, &shop)
}

func update(c *gin.Context, shop *ShopRequest) {
	//1.更新数据库
	//当通过 struct 更新时，GORM 只会更新非零字段。
	//若想确保指定字段被更新,应使用Select更新选定字段，或使用map来完成更新
	data := shop.ToModel()
	tbshop := query.TbShop
	_, err := tbshop.Where(tbshop.ID.Eq(shop.ID)).Updates(data)
	if err != nil {
		slog.Error("update mysql bad", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	//2.删除缓存
	key := shopKeyPrefix + strconv.Itoa(int(shop.ID))
	db.RedisDb.Del(context.Background(), key)

	response.Success(c, nil)
}

// 添加商铺
// post /api/v1/shop
func AddShop(c *gin.Context) {
	var shop ShopRequest
	err := c.ShouldBindJSON(&shop)
	if err != nil {
		slog.Error("bindjson bad", "err", err)
		response.Error(c, response.ErrBind)
		return
	}

	// 添加时忽略ID字段
	shop.ID = 0

	data := shop.ToModel()
	err = query.TbShop.Create(data)
	if err != nil {
		slog.Error("mysql create shop err", "err", err)
		response.Error(c, response.ErrDatabase)
		return
	}

	response.Success(c, gin.H{"id": data.ID})
}

// 删除商铺
// delete /api/v1/shop/:shopId
func DelShop(c *gin.Context) {
	id := c.Param("shopId")
	if id == "" {
		response.Error(c, response.ErrValidation, "id is null")
		return
	}

	val, err := strconv.ParseInt(id, 10, 64)
	if err != nil || val <= 0 {
		response.Error(c, response.ErrValidation, "invalid shop id")
		return
	}
	shop := query.TbShop
	result, err := shop.Where(shop.ID.Eq(uint64(val))).Delete()
	if err != nil {
		response.Error(c, response.ErrDatabase)
	}
	if result.RowsAffected == 0 {
		response.Error(c, response.ErrNotFound, "shop not found")
		return
	}
	//删除缓存
	key := "cache:shop:type:sortBy:empty:current:1"
	_, err = db.RedisDb.Del(context.Background(), key).Result()
	if err != nil {
		slog.Error("redis delete shop error", "error", err)
	}

	response.Success(c, nil)
}
