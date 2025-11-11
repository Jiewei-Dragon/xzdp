package Shop

import (
	"context"
	"strconv"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"

	"github.com/bytedance/sonic"
)

func getShopsByTypeIdFromDB(idInt int, sortBy string, current int) ([]*model.TbShop, error) {
	shopsQuery := query.TbShop
	offset := (current - 1) * ShopPageSize // 跳过的记录数
	limit := ShopPageSize                  // 每页返回的记录数
	if sortBy == "comments" {
		return shopsQuery.Where(shopsQuery.TypeID.Eq(uint64(idInt))).Order(shopsQuery.Comments.Desc()).Offset(offset).Limit(limit).Find()
	}
	if sortBy == "score" {
		return shopsQuery.Where(shopsQuery.TypeID.Eq(uint64(idInt))).Order(shopsQuery.Score.Desc()).Offset(offset).Limit(limit).Find()
	}
	return shopsQuery.Where(shopsQuery.TypeID.Eq(uint64(idInt))).Offset(offset).Limit(limit).Find()
}

func setShopsByTypeIdToCache(CacheKey string, Shops []*model.TbShop) error {
	//	因为数据量较小，并且访问比较集中，所以采用缓存分页数据方案
	//	cache:shop:typeId:{typeId}:sort:{sortBy}:page:{current}，即把每种排序的每一页的都缓存
	//  1.先将数据序列化为 JSON
	jsonData, err := sonic.Marshal(Shops)
	if err != nil {
		return err
	}
	return db.RedisDb.Set(context.Background(), CacheKey, jsonData, shopCacheTTL).Err()
}

func getShopsByTypeIdFromCache(CacheKey string) ([]*model.TbShop, error) {
	res, err := db.RedisDb.Get(context.Background(), CacheKey).Result()
	if err != nil {
		return nil, err
	}
	var shopRes []*model.TbShop
	er := sonic.Unmarshal([]byte(res), &shopRes)
	if er != nil {
		return nil, er
	}
	return shopRes, nil
}

// 函数成功执行时，返回一个指向 model.TbShop 结构体的指针
// 表示函数执行过程中可能出现的错误。如果执行成功，error 为 nil
func getShopByIdFromDB(idInt int) (*model.TbShop, error) {
	shopQuery := query.TbShop
	return shopQuery.Where(shopQuery.ID.Eq(uint64(idInt))).First()
}

func getShopByIdFromCache(CacheKey string) (*model.TbShop, error) {
	res, err := db.RedisDb.Get(context.Background(), CacheKey).Result()
	if res == "" || err != nil {
		return nil, err
	} else {
		var shop *model.TbShop
		err = sonic.Unmarshal([]byte(res), &shop)
		if err != nil {
			return nil, err
		}
		return shop, nil
	}
}

func setShopByIdtoCache(CacheKey string, shop model.TbShop) error {
	shopJson, err := sonic.Marshal(shop)
	if shopJson == nil || err != nil {
		return err
	}
	return db.RedisDb.Set(context.Background(), CacheKey, shopJson, shopCacheTTL).Err()
}

func getShopTypeListFromDB() ([]*model.TbShopType, error) {
	shopTypeQuery := query.TbShopType
	return shopTypeQuery.Order(shopTypeQuery.Sort).Find()
}

func setShopTypeListToCache(shopTypeList []*model.TbShopType) error {
	b, err := sonic.Marshal(shopTypeList)
	if b == nil || err != nil {
		return err
	}
	// 使用 Set 命令存储为 String 类型，并设置过期时间
	return db.RedisDb.Set(context.Background(), shopKeyPrefix+shopTypeKey+":list", b, shopTypeCacheTTL).Err()
}

// sonic.Marshal：参数是任何数据类型，返回json字节类型的数据，Go 对象 → JSON 字节。
// sonic.Unmarshal：第一个参数是json字节，第二个参数是目标Go对象的指针，JSON 字节 → Go 对象。
func getShopTypeListFromCache() (*[]model.TbShopType, error) {
	// 使用 Get 命令读取 String 类型的数据，key 与写入时保持一致
	res, err := db.RedisDb.Get(context.Background(), shopKeyPrefix+shopTypeKey+":list").Result()
	if res == "" || err != nil {
		return nil, err
	}

	var types []model.TbShopType
	// 直接将 JSON 字符串反序列化为 []model.TbShopType
	err = sonic.Unmarshal([]byte(res), &types)
	if err != nil {
		return nil, err
	}
	return &types, nil
}

func setHotBlogToCache(TbBlogList []*model.TbBlog) error {
	var TbBlogListJson []interface{}
	for _, blog := range TbBlogList {
		b, _ := sonic.Marshal(blog)
		TbBlogListJson = append(TbBlogListJson, string(b))
	}
	err := db.RedisDb.LPush(context.Background(), BlogPrefix+HotBlogKey, TbBlogListJson).Err()
	if err != nil {
		return err
	}
	return db.RedisDb.Expire(context.Background(), BlogPrefix+HotBlogKey, HotBlogTTL).Err()
}

func getBlogByPageNumFromCache(pageNum string) ([]*model.TbBlog, error) {
	Num, err := strconv.Atoi(pageNum)
	if err != nil {
		return nil, err
	}

	pageSize := 10
	startNum := (Num - 1) * pageSize
	endNum := startNum + pageSize - 1

	// 使用 LRange 读取指定范围的博客
	res, err := db.RedisDb.LRange(context.Background(), BlogPrefix+HotBlogKey, int64(startNum), int64(endNum)).Result()
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, nil // 缓存未命中或该页无数据
	}

	// 反序列化每个博客
	blogs := make([]*model.TbBlog, 0, len(res))
	for _, item := range res {
		var blog model.TbBlog
		err = sonic.Unmarshal([]byte(item), &blog)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, &blog)
	}

	return blogs, nil
}

func getBlogByPageNumFromDB(pageNum string) ([]*model.TbBlog, error) {
	Num, err := strconv.Atoi(pageNum)
	if err != nil {
		return nil, err
	}
	pageSize := 10 // 每页10条
	offset := (Num - 1) * pageSize
	TbBlogQuery := query.TbBlog
	return TbBlogQuery.Order(TbBlogQuery.Liked.Desc()).Offset(offset).Limit(pageSize).Find()
}
