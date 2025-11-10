package router

import (
	"net/http"
	"path/filepath"
	"xzdp/handle/ShopService"
	"xzdp/handle/UserService"
	"xzdp/middleware"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
)

type User struct {
	Name string `json:"name" binding:"required"`
	Age  int    `json:"age" `
}

func HandleNotFound(c *gin.Context) {
	response.Error(c, response.ErrNotFound, "route not found")
}

func NewRouter() *gin.Engine {
	//gin.SetMode(gin.ReleaseMode) //将项目设为开发模式，减少输出的log，提高性能
	r := gin.Default()

	// 配置静态文件服务 - 提供静态资源（CSS、JS、图片等）
	staticDir := filepath.Join("nginx-1.18.0", "html", "hmdp")

	// 静态资源目录
	r.Static("/css", filepath.Join(staticDir, "css"))
	r.Static("/js", filepath.Join(staticDir, "js"))
	r.Static("/imgs", filepath.Join(staticDir, "imgs"))

	// favicon
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))

	// 为了支持相对路径（如 ./css/index.css），也需要提供静态文件服务
	r.StaticFS("/static", http.Dir(staticDir))

	// 根路径返回首页（放在最前面，确保优先匹配）
	r.GET("/", func(c *gin.Context) {
		indexPath := filepath.Join(staticDir, "index.html")
		c.File(indexPath)
	})

	// Use为当前路由组中的所有路由绑定中间件，使得该组内的所有请求在到达具体处理函数前，都会先经过这些中间件的处理
	public := r.Group("/api")
	public.Use(middleware.OptionalJWT())
	{
		public.GET("/shop/:id", ShopService.QueryShopById)
		public.GET("/shop-type/list", ShopService.QueryShopTypeList)
		public.GET("/shop/of/type", ShopService.GetShopByTypeId)
		public.POST("/user/code", UserService.SendVerifyCode)
		public.POST("/user/login", UserService.Login)
		public.GET("/blog/hot", ShopService.GetHotBlog)
	}
	auth := r.Group("/api")
	auth.Use(middleware.OptionalJWT(), middleware.RequireAuth())
	{
		//登录时第一次获取用户信息
		auth.GET("/user/me", UserService.GetUserInfo)
		////每次点击个人信息页时获取用户信息
		auth.GET("/user/info/:userId", UserService.GetUserInfoById)
		auth.POST("/user/logout", UserService.Logout)
		auth.PUT("user/nickname", UserService.EditNickname)
	}
	r.StaticFile("/index.html", filepath.Join(staticDir, "index.html"))
	r.StaticFile("/login.html", filepath.Join(staticDir, "login.html"))
	r.StaticFile("/shop-list.html", filepath.Join(staticDir, "shop-list.html"))
	r.StaticFile("/shop-detail.html", filepath.Join(staticDir, "shop-detail.html"))
	r.StaticFile("/blog-detail.html", filepath.Join(staticDir, "blog-detail.html"))
	r.StaticFile("/blog-edit.html", filepath.Join(staticDir, "blog-edit.html"))
	r.StaticFile("/info.html", filepath.Join(staticDir, "info.html"))
	r.StaticFile("/info-edit.html", filepath.Join(staticDir, "info-edit.html"))
	r.StaticFile("/other-info.html", filepath.Join(staticDir, "other-info.html"))
	// 404处理 - 必须放在最后
	r.NoRoute(HandleNotFound)
	return r
}
