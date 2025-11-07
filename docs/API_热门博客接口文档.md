# 热门博客接口文档

## 接口信息

**接口名称：** 查询热门博客列表（分页）

**接口路径：** `GET /api/blog/hot`

**是否需要认证：** 否（公开接口）

**接口描述：** 获取热门博客列表，支持分页查询。用于首页展示热门博客内容，支持下拉加载更多。

---

## 请求参数

### Query 参数

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| current | int | 是 | 1 | 当前页码，从 1 开始 |

### 请求示例

```http
GET /api/blog/hot?current=1
```

---

## 响应数据

### 响应格式

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "shopId": 1,
      "userId": 1010,
      "title": "探店标题",
      "images": "图片1.jpg,图片2.jpg,图片3.jpg",
      "content": "探店的文字描述内容",
      "liked": 128,
      "comments": 15,
      "createTime": "2024-01-01T10:00:00",
      "updateTime": "2024-01-01T10:00:00",
      "name": "用户昵称",
      "icon": "/imgs/icons/user1.png",
      "isLike": false
    }
  ]
}
```

### 响应字段说明

#### 外层字段

| 字段名 | 类型 | 说明 |
|--------|------|------|
| success | boolean | 请求是否成功 |
| data | array | 博客列表数组 |

#### data 数组中的博客对象字段

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | uint64 | 是 | 博客ID（主键） |
| shopId | int64 | 是 | 关联的商户ID |
| userId | uint64 | 是 | 发布博客的用户ID |
| title | string | 是 | 博客标题 |
| images | string | 是 | 图片URL列表，多张图片用英文逗号分隔，最多9张 |
| content | string | 是 | 博客正文内容 |
| liked | uint32 | 是 | 点赞数量 |
| comments | uint32 | 是 | 评论数量 |
| createTime | string | 是 | 创建时间（ISO 8601 格式） |
| updateTime | string | 是 | 更新时间（ISO 8601 格式） |
| name | string | 是 | 发布者昵称（需要关联用户表查询） |
| icon | string | 是 | 发布者头像URL（需要关联用户表查询） |
| isLike | boolean | 是 | 当前用户是否已点赞（需要判断当前登录用户） |

---

## 前端处理逻辑

前端收到响应后会进行以下处理：

```javascript
// 1. 提取第一张图片作为封面
data.forEach(b => b.img = b.images.split(",")[0]);

// 2. 追加到已有列表（支持滚动加载）
this.blogs = this.blogs.concat(data);
```

---

## 业务逻辑要求

### 1. 分页规则

- **每页数量：** 建议 10-20 条（需根据实际情况定义）
- **排序规则：** 按热度排序（推荐算法：点赞数 + 评论数，或按创建时间倒序）
- **边界处理：** 最后一页返回空数组 `[]`

### 2. 数据关联

需要关联查询以下数据：

```
博客表 (tb_blog)
  ├─ 关联用户表 (tb_user) 获取：
  │   ├─ name (昵称)
  │   └─ icon (头像)
  │
  └─ 判断点赞状态：
      └─ 查询点赞表判断当前用户是否已点赞该博客
```

### 3. 点赞状态判断

- **已登录用户：** 查询点赞表，判断该用户是否点赞过该博客
- **未登录用户：** `isLike` 统一返回 `false`

### 4. 图片处理

- `images` 字段存储格式：`"img1.jpg,img2.jpg,img3.jpg"`
- 前端会自动提取第一张图片作为封面显示

---

## 数据库表结构参考

### tb_blog 表（博客表）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint | 商户ID |
| user_id | bigint unsigned | 用户ID |
| title | varchar(255) | 标题 |
| images | varchar(2048) | 图片列表（逗号分隔） |
| content | varchar(2048) | 内容 |
| liked | int unsigned | 点赞数 |
| comments | int unsigned | 评论数 |
| create_time | timestamp | 创建时间 |
| update_time | timestamp | 更新时间 |

### 需要关联的表

- `tb_user` - 用户表（获取昵称和头像）
- `tb_blog_like` - 点赞表（判断用户是否点赞，可能需要创建）

---

## 实现建议

### 1. SQL 查询示例（伪代码）

```sql
-- 分页查询热门博客
SELECT 
    b.id, b.shop_id, b.user_id, b.title, b.images, 
    b.content, b.liked, b.comments, b.create_time, b.update_time,
    u.nick_name as name, u.icon
FROM tb_blog b
LEFT JOIN tb_user u ON b.user_id = u.id
ORDER BY (b.liked + b.comments) DESC  -- 按热度排序
LIMIT ? OFFSET ?;  -- 分页参数
```

### 2. 分页参数计算

```go
pageSize := 10  // 每页数量
offset := (current - 1) * pageSize
limit := pageSize
```

### 3. 点赞状态查询

```go
// 如果用户已登录
if userId > 0 {
    // 批量查询当前用户对这些博客的点赞状态
    // SELECT blog_id FROM tb_blog_like 
    // WHERE user_id = ? AND blog_id IN (?, ?, ...)
}
```

---

## 错误处理

### 常见错误

| 错误码 | 说明 | 处理方式 |
|--------|------|----------|
| 400 | 参数错误（current 不是数字） | 返回参数错误提示 |
| 500 | 数据库查询失败 | 返回服务器错误 |

### 错误响应格式

```json
{
  "success": false,
  "errorMsg": "错误信息描述"
}
```

---

## 测试用例

### 测试场景 1：首页加载

```
请求：GET /api/blog/hot?current=1
预期：返回第一页博客列表（10条）
```

### 测试场景 2：下拉加载更多

```
请求：GET /api/blog/hot?current=2
预期：返回第二页博客列表
```

### 测试场景 3：超出页数范围

```
请求：GET /api/blog/hot?current=999
预期：返回空数组 []
```

### 测试场景 4：已登录用户的点赞状态

```
请求：GET /api/blog/hot?current=1
请求头：Authorization: Bearer {token}
预期：isLike 字段根据实际点赞情况返回 true/false
```

---

## 实现步骤建议

1. **第一步：** 在 `handle/BlogService/` 目录下创建 `BlogService.go`
2. **第二步：** 实现基础的分页查询（不含用户信息）
3. **第三步：** 添加用户信息关联查询（name, icon）
4. **第四步：** 添加点赞状态判断逻辑
5. **第五步：** 测试各种场景

---

## 注意事项

⚠️ **重要提示：**

1. 前端会自动拼接 `/api` 前缀，路由中配置为 `/blog/hot` 即可
2. 响应必须包含 `success` 字段，前端会以此判断请求是否成功
3. 图片字段 `images` 是逗号分隔的字符串，不是数组
4. 点赞状态 `isLike` 必须返回，未登录时返回 `false`
5. 当没有更多数据时，返回空数组 `[]` 而不是 `null`

---

## 相关接口

- `GET /api/blog/:id` - 查询单个博客详情
- `PUT /api/blog/like/:id` - 点赞/取消点赞博客

---

**文档版本：** v1.0  
**创建时间：** 2024-01-07  
**维护人员：** Backend Team