package User

import (
	"context"
	"strconv"
	"time"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"
	"xzdp/pkg/response"

	"github.com/bytedance/sonic"
)

func getUserByIdFromDb(id int64) (*model.TbUser, error) {
	userQuery := query.TbUser
	return userQuery.Where(userQuery.ID.Eq(uint64(id))).First()
}

func UpdateUserInfoById(user *model.TbUser) error {
	userQuery := query.TbUser
	return userQuery.Save(user)
}

func getUserByIdFromCache(id string) (*model.TbUser, error) {
	res, err := db.RedisDb.Get(context.Background(), userPrefix+inforKeyPrefix+":"+id).Result()
	if res == "" || err != nil {
		return nil, response.NewBusinessError(response.ErrExpired, "用户信息不存在或已过期")
	}
	var user model.TbUser
	err = sonic.Unmarshal([]byte(res), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func setUserToCache(user *model.TbUser) error {
	b, err := sonic.Marshal(user)
	if err != nil {
		return err
	}
	return db.RedisDb.Set(context.Background(),
		userPrefix+inforKeyPrefix+":"+strconv.FormatUint(user.ID, 10),
		string(b), time.Duration(userInfoCacheTTL)).Err()
}

func deleteUserInfoFromCache(id string) error {
	return db.RedisDb.Del(context.Background(), userPrefix+inforKeyPrefix+":"+id).Err()
}

// 处理成脱敏手机号134****3310
func MaskPhoneNumber(phone string) string {
	return phone[:3] + "****" + phone[7:]
}
