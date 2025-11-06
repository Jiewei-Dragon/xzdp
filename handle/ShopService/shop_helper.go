package ShopService

import (
	"context"
	"fmt"
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
	//调用 Err() 用于获取命令执行过程中可能出现的错误
	return db.RedisDb.LPush(context.Background(), shopKeyPrefix+shopTypeKey+":list", string(b)).Err()
}

func getShopTypeListFromCache() (*model.TbShopType, error) {
	res, err := db.RedisDb.LRange(context.Background(), shopKeyPrefix+shopTypeKey+":list", 0, -1).Result()
	if res == nil || err != nil {
		fmt.Print("商户类型Cache未命中")
		return nil, err
	}
	var types model.TbShopType
	//将Redis返回的 JSON 字符串（res[0]）解析为 Go 中的结构体对象（types）。
	err = sonic.Unmarshal([]byte(res[0]), &types)
	if err != nil {
		return nil, err
	}
	return &types, nil
}
