# XZDP - 小众点评

点评系统后端服务，基于 Go + Gin 框架开发。

## 项目结构

```
xzdp/
├── config/          # 配置管理
├── configs/         # 配置文件
├── dal/             # 数据访问层（ORM）
├── db/              # 数据库连接
├── handle/          # 业务处理层
│   ├── ShopService/ # 商户服务
│   └── UserService/ # 用户服务
├── middleware/      # 中间件（JWT认证等）
├── pkg/             # 公共包
│   ├── logger/      # 日志
│   └── response/    # 响应处理
├── router/          # 路由配置
├── scripts/         # 脚本
└── nginx-1.18.0/    # Nginx静态文件服务
```

## 功能特性

- ✅ 用户登录注册（手机号+验证码）
- ✅ JWT 认证授权
- ✅ 商户信息查询
- ✅ 商户类型列表
- ✅ 用户信息管理
- ✅ Redis 缓存支持
- ✅ MySQL 数据库支持

## 技术栈

- Go 1.21+
- Gin Web 框架
- GORM ORM
- Redis
- MySQL
- JWT 认证

## 配置说明

配置文件路径：`configs/config.yaml`

需要配置：
- MySQL 数据库连接
- Redis 连接
- JWT 密钥和过期时间
- 服务器端口

## 运行

```bash
go run main.go
```

或

```bash
go build -o xzdp.exe
./xzdp.exe
```

## API 文档

### 公开接口（无需认证）

- `POST /api/user/code` - 发送验证码
- `POST /api/user/login` - 用户登录
- `GET /api/shop/:id` - 查询商户详情
- `GET /api/shop-type/list` - 查询商户类型列表

### 需要认证的接口

- `GET /api/user/me` - 获取当前用户信息

## License

MIT

