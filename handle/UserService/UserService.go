package UserService

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"regexp"
	"strconv"
	"time"
	"xzdp/dal/model"
	"xzdp/dal/query"
	"xzdp/db"
	"xzdp/middleware"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// loginReq登录请求结构体（统一使用 phone 作为用户标识）
type loginReqstruct struct {
	Phone    string `json:"phone" binding:"required"`                  // 手机号（必填）
	Password string `json:"password" binding:"omitempty,min=6,max=20"` // 密码（仅密码登录时用）
	Code     string `json:"code" binding:"omitempty,min=4,max=6"`      // 验证码（仅验证码登录时用）
}

type userResponse struct {
	ID       uint64 `json:"id"`
	Phone    string `json:"phone"`
	NickName string `json:"nickName"` // 使用驼峰命名，匹配前端
	Icon     string `json:"icon"`
}

// 修改昵称请求结构体
type updateNickNameReq struct {
	NickName string `json:"nickName" binding:"required,min=1,max=15"` // 昵称，1-20个字符
}

const (
	userPrefix       = "cache:user"
	phoneKeyPrefix   = ":phone"
	inforKeyPrefix   = ":info"
	codeExpiration   = 3 * time.Minute
	userInfoCacheTTL = 10 * time.Minute
)

var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

func isValidPhone(phone string) bool {
	if phoneRe.MatchString(phone) {
		return true
	} else {
		return false
	}
}

// http://localhost:8080/api/user/code/:phone
func SendVerifyCode(c *gin.Context) {
	//1.手机号格式校验
	phoneNum := c.Query("phone")
	if !isValidPhone(phoneNum) {
		response.Error(c, response.ErrValidation, "手机号格式有误！")
		return
	}
	//2.发送验证码，一手一码（典型键值对，还有过期时间-->存在redis中）
	rand.Seed(time.Now().UnixNano())
	code := 1000 + rand.Intn(9000)
	db.RedisDb.SetNX(c, userPrefix+phoneKeyPrefix+":"+phoneNum, code, codeExpiration)
	response.Success(c, gin.H{"code": code})
}

func Login(c *gin.Context) {
	//1.手机号格式校验
	var loginRequest loginReqstruct
	err := c.ShouldBindJSON(&loginRequest)
	if err != nil {
		response.Error(c, response.ErrInvalidJSON)
		return
	}
	if !isValidPhone(loginRequest.Phone) {
		response.Error(c, response.ErrValidation, "手机号格式有误！")
		return
	}
	//2.从redis拿验证码进行校验
	var user *model.TbUser
	user, e := CodeLogin(loginRequest)
	if user == nil || e != nil {
		response.HandleBusinessError(c, e)
		return
	}
	//3. 生成Token，需要用到手机号 + userId
	Token, err := middleware.GenerateToken(loginRequest.Phone, int64(user.ID))
	if err != nil {
		slog.Error("生成Token失败", "err", err)
		response.Error(c, response.ErrorLoginFaild, "")
		return
	}
	response.Success(c, gin.H{
		"token": Token,
		"user": userResponse{ //返回给前端展示或者缓存的用户信息
			ID:       user.ID,
			Phone:    user.Phone,
			NickName: MaskPhoneNumber(user.Phone),
			Icon:     user.Icon,
		},
	})
}

func CodeLogin(loginReqstruct loginReqstruct) (*model.TbUser, error) {
	// 1.去缓存找对应的验证码
	DbCode, err := db.RedisDb.Get(context.Background(), userPrefix+phoneKeyPrefix+":"+loginReqstruct.Phone).Result()
	if DbCode == "" || err != nil {
		return nil, response.NewBusinessError(response.ErrExpired, "验证码不存在或已过期")
	}
	if loginReqstruct.Code != DbCode {
		return nil, response.NewBusinessError(response.ErrPasswordIncorrect, "验证码错误")
	}
	//2. 验证码正确，删除Redis中的验证码，并生成Token
	db.RedisDb.Del(context.Background(), userPrefix+phoneKeyPrefix+":"+loginReqstruct.Phone)
	//3.判断是否是新用户，是则新建帐号
	userQuery := query.TbUser
	user, err := userQuery.Where(userQuery.Phone.Eq(loginReqstruct.Phone)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			newUser := &model.TbUser{Phone: loginReqstruct.Phone, NickName: MaskPhoneNumber(loginReqstruct.Phone)}
			err = userQuery.Create(newUser)
			if err != nil {
				return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
			}
			user = newUser
		} else {
			// 其他数据库错误
			return nil, response.WrapBusinessError(response.ErrDatabase, err, "")
		}
	}
	return user, nil
}

func GetUserInfo(c *gin.Context) {
	//1. 从上下文获取用户信息
	userId := c.GetInt64(middleware.CtxKeyUserId)
	//2. 从Cache获取用户信息
	user, err := getUserByIdFromCache(strconv.FormatInt(userId, 10))
	//3. 没有就去DB找并写回Cache再返回
	if user == nil || err != nil {
		user, err = getUserByIdFromDb(userId)
		if err != nil {
			response.HandleBusinessError(c, err)
			return
		}
		err = setUserToCache(user)
		if err != nil {
			// 缓存设置失败不影响返回，只记录日志
			slog.Error("设置用户缓存失败", "err", err)
		}
	}
	//4. 转换为响应格式（使用驼峰命名，匹配前端）
	response.Success(c, userResponse{
		ID:       user.ID,
		Phone:    user.Phone,
		NickName: user.NickName,
		Icon:     user.Icon,
	})
}

func GetUserInfoById(c *gin.Context) {
	userId := c.Param("userId")
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		response.HandleBusinessError(c, err) // 假设有定义
		return
	}
	user, err := getUserByIdFromCache(strconv.FormatInt(id, 10))
	if user == nil || err != nil {
		user, err = getUserByIdFromDb(id)
		if err != nil {
			response.HandleBusinessError(c, err)
			return
		}
		err = setUserToCache(user)
		if err != nil {
			slog.Error("设置用户缓存失败", "err", err)
		}
	}
	response.Success(c, userResponse{
		ID:       user.ID,
		Phone:    user.Phone,
		NickName: user.NickName,
		Icon:     user.Icon,
	})
}

func EditNickname(c *gin.Context) {
	//1. 从上下文中获取用户ID
	userId := c.GetInt64(middleware.CtxKeyUserId)

	//2. 绑定请求参数（使用专门的请求结构体，匹配前端的 nickName 字段）
	var req updateNickNameReq
	//2.1 从请求体中获取新昵称
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.Error(c, response.ErrValidation, "请求参数格式错误")
		return
	}
	if len(req.NickName) < 1 || len(req.NickName) > 15 {
		response.Error(c, response.ErrValidation, "昵称长度不合规，请重试")
		return
	}

	//4. 更新数据库
	user := &model.TbUser{
		ID:       uint64(userId),
		NickName: req.NickName,
		Phone:    c.GetString(middleware.CtxKeyUserPhone),
	}
	err = UpdateUserInfoById(user)
	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}
	deleteUserInfoFromCache(strconv.FormatInt(userId, 10))
	response.Success(c, gin.H{"message": "昵称修改成功"})
}

func Logout(c *gin.Context) {
	//1.拿到唯一标识
	userId := c.GetInt64(middleware.CtxKeyUserId)
	//2.删除缓存
	err := deleteUserInfoFromCache(strconv.FormatInt(userId, 10))
	if err != nil {
		response.HandleBusinessError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "退出成功"})
}
