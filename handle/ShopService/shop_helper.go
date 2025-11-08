package ShopService

import (
	"context"
	"strconv"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"

	"github.com/bytedance/sonic"
)

func getShopByIDFromDB(idInt int) (*model.TbShop, error) {
	shopQuery := query.TbShop
	return shopQuery.Where(shopQuery.ID.Eq(uint64(idInt))).First()
}

// 函数成功执行时，返回一个指向 model.TbShop 结构体的指针
// 表示函数执行过程中可能出现的错误。如果执行成功，error 为 nil
func getShopFromCache(idStr string) (*model.TbShop, error) {
	res, err := db.RedisDb.Get(context.Background(), idStr).Result()
	if res == "" || err != nil {
		return nil, err
	} else {
		var shop model.TbShop
		//将JSON格式的字节流转换为 Go 中的结构体，返回值是一个error，解析成功时为nil，失败返回具体结果
		err = sonic.Unmarshal([]byte(res), &shop)
		if err != nil {
			return nil, err
		}
		return &shop, nil
	}
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
	if len(TbBlogList) == 0 {
		return nil // 空数组不需要缓存
	}

	ctx := context.Background()
	key := shopKeyPrefix + HotBlogKey

	// 使用管道批量写入，提高性能
	pipe := db.RedisDb.Pipeline()

	// 先删除旧数据
	pipe.Del(ctx, key)

	// 将每个博客序列化后添加到 List 中
	for _, blog := range TbBlogList {
		b, err := sonic.Marshal(blog)
		if err != nil {
			return err
		}
		pipe.RPush(ctx, key, string(b))
	}

	// 设置过期时间
	pipe.Expire(ctx, key, HotBlogTTL)

	// 执行管道命令
	_, err := pipe.Exec(ctx)
	return err
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
	res, err := db.RedisDb.LRange(context.Background(), shopKeyPrefix+HotBlogKey, int64(startNum), int64(endNum)).Result()
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
	return TbBlogQuery.Order(TbBlogQuery.Liked).Offset(offset).Limit(pageSize).Find()
}
