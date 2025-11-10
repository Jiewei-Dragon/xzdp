# 商铺类型查询接口文档

## 接口信息

**接口路径**: `/api/shop/of/type`

**请求方法**: `GET`

**接口描述**: 根据商铺类型ID查询该类型下的商铺列表，支持分页、排序和地理位置距离计算

---

## 请求参数

### Query Parameters (URL参数)

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| typeId | int | 是 | 商铺类型ID | 1 |
| current | int | 是 | 当前页码，从1开始 | 1 |
| x | float64 | 是 | 用户当前位置的经度 | 120.149993 |
| y | float64 | 是 | 用户当前位置的纬度 | 30.334229 |
| sortBy | string | 否 | 排序字段，可选值：<br/>- "" (空字符串): 按距离排序<br/>- "comments": 按人气(评论数)排序<br/>- "score": 按评分排序 | "comments" |

### 请求示例

```
GET /api/shop/of/type?typeId=1&current=1&x=120.149993&y=30.334229
GET /api/shop/of/type?typeId=1&current=1&x=120.149993&y=30.334229&sortBy=comments
GET /api/shop/of/type?typeId=1&current=2&x=120.149993&y=30.334229&sortBy=score
```

---

## 响应数据

### 成功响应

**HTTP状态码**: `200 OK`

**响应格式**: `application/json`

**响应体**: 返回商铺对象数组

```json
[
  {
    "id": 1,
    "name": "商铺名称",
    "typeId": 1,
    "images": "图片URL1,图片URL2,图片URL3",
    "area": "所在区域",
    "address": "详细地址",
    "avgPrice": 88,
    "score": 45,
    "comments": 120,
    "distance": 1.5
  },
  {
    "id": 2,
    "name": "另一家商铺",
    "typeId": 1,
    "images": "图片URL",
    "area": "所在区域",
    "address": "详细地址",
    "avgPrice": 128,
    "score": 48,
    "comments": 200,
    "distance": 2.3
  }
]
```

### 响应字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int64 | 商铺ID |
| name | string | 商铺名称 |
| typeId | int64 | 商铺类型ID |
| images | string | 商铺图片URLs，多个图片用逗号分隔 |
| area | string | 商铺所在区域 |
| address | string | 商铺详细地址 |
| avgPrice | int64 | 人均消费价格（单位：元） |
| score | int32 | 商铺评分（实际评分需要除以10，如45表示4.5分） |
| comments | int32 | 评论数/人气值 |
| distance | float64 | 距离用户的距离（单位：公里） |

### 错误响应

**HTTP状态码**: `400 Bad Request` / `401 Unauthorized` / `500 Internal Server Error`

```json
{
  "error": "错误信息描述"
}
```

---

## 业务逻辑说明

### 1. 分页逻辑
- 前端首次加载时 `current=1`
- 用户滚动到底部时，`current` 自动递增（`current++`）
- 后端需要实现分页查询，建议每页返回固定数量的商铺（如10条）

### 2. 排序逻辑
- **默认排序（sortBy为空）**: 按照距离从近到远排序
- **人气排序（sortBy=comments）**: 按照评论数降序排序
- **评分排序（sortBy=score）**: 按照评分降序排序

### 3. 距离计算
- 前端传入用户当前位置的经纬度 (x, y)
- 后端需要根据商铺的经纬度和用户位置计算距离
- 距离单位为公里（km）
- 建议使用 Haversine 公式或数据库的地理位置函数计算距离

### 4. 前端处理逻辑
- 前端接收到商铺列表后，会将多张图片字符串拆分，只取第一张图片展示
- 评分会除以10显示（如 score=45 显示为 4.5星）
- 距离直接显示为 "X.X km"

---

## 数据库表结构参考

### tb_shop 表字段（推测）

```sql
- id: 商铺ID (主键)
- name: 商铺名称
- type_id: 商铺类型ID (外键)
- images: 图片URLs
- area: 区域
- address: 地址
- x: 经度
- y: 纬度
- avg_price: 人均价格
- score: 评分
- comments: 评论数
```

---

## 实现要点

### 后端开发注意事项：

1. **参数验证**
   - typeId 必须大于0
   - current 必须大于0
   - x, y 必须是有效的经纬度范围
   - sortBy 只能是空字符串、"comments" 或 "score"

2. **距离计算**
   - 需要根据商铺表中的经纬度字段和用户传入的经纬度计算距离
   - 如果使用MySQL，可以使用 `ST_Distance_Sphere` 函数
   - 如果使用Redis GEO，可以使用 `GEORADIUS` 命令

3. **性能优化**
   - 考虑对查询结果进行缓存（特别是热门类型）
   - 可以使用 Redis 的 GEO 数据结构存储商铺位置信息
   - 分页查询时注意索引优化

4. **数据一致性**
   - 确保返回的商铺都属于指定的 typeId
   - 确保所有商铺都有有效的地理位置信息

---

## 测试用例

### 测试场景1: 基本查询
```
请求: GET /api/shop/of/type?typeId=1&current=1&x=120.149993&y=30.334229
期望: 返回类型为1的商铺列表，按距离排序
```

### 测试场景2: 按人气排序
```
请求: GET /api/shop/of/type?typeId=1&current=1&x=120.149993&y=30.334229&sortBy=comments
期望: 返回类型为1的商铺列表，按评论数降序
```

### 测试场景3: 按评分排序
```
请求: GET /api/shop/of/type?typeId=1&current=1&x=120.149993&y=30.334229&sortBy=score
期望: 返回类型为1的商铺列表，按评分降序
```

### 测试场景4: 分页查询
```
请求: GET /api/shop/of/type?typeId=1&current=2&x=120.149993&y=30.334229
期望: 返回第2页的商铺列表
```

### 测试场景5: 无效参数
```
请求: GET /api/shop/of/type?typeId=0&current=1&x=120.149993&y=30.334229
期望: 返回400错误
```

---

## 相关接口

- `GET /shop-type/list` - 获取商铺类型列表
- `GET /user/me` - 获取当前登录用户信息
- `GET /shop-detail.html?id={id}` - 商铺详情页面

---

**文档版本**: v1.0  
**创建时间**: 2025-11-08  
**维护人员**: Backend Team