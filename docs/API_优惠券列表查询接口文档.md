# 优惠券列表查询接口文档

## 接口信息

**接口路径**: `/api/voucher/list/:shopId`

**请求方法**: `GET`

**接口描述**: 根据商铺ID查询该商铺的所有优惠券列表，包括普通券和秒杀券信息

**认证要求**: 需要用户登录，需要在请求头中携带有效的 token

---

## 请求参数

### Path Parameters (路径参数)

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| shopId | int64 | 是 | 商铺ID | 1 |

### 请求示例

```
GET /api/voucher/list/1
```

### 请求头

```
Authorization: Bearer <token>
或
token: <token>
```

---

## 响应数据

### 成功响应

**HTTP状态码**: `200 OK`

**响应格式**: `application/json`

**响应体**: 返回优惠券对象数组

```json
[
  {
    "id": 1,
    "shop_id": 1,
    "title": "100元代金券",
    "sub_title": "周一至周五可用",
    "rules": "单笔消费满200元可用",
    "pay_value": 10000,
    "actual_value": 20000,
    "type": 0,
    "status": 1,
    "create_time": "2024-01-01T10:00:00Z",
    "update_time": "2024-01-01T10:00:00Z"
  },
  {
    "id": 2,
    "shop_id": 1,
    "title": "50元秒杀券",
    "sub_title": "限时抢购",
    "rules": "单笔消费满100元可用",
    "pay_value": 5000,
    "actual_value": 10000,
    "type": 1,
    "status": 1,
    "create_time": "2024-01-01T10:00:00Z",
    "update_time": "2024-01-01T10:00:00Z",
    "stock": 100,
    "begin_time": "2024-01-10T10:00:00Z",
    "end_time": "2024-01-10T12:00:00Z"
  }
]
```

### 响应字段说明

#### 基础字段（所有优惠券）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | uint64 | 优惠券ID |
| shop_id | uint64 | 商铺ID |
| title | string | 代金券标题 |
| sub_title | string | 副标题 |
| rules | string | 使用规则 |
| pay_value | uint64 | 支付金额，单位是分。例如10000代表100元 |
| actual_value | int64 | 抵扣金额，单位是分。例如20000代表200元 |
| type | uint32 | 优惠券类型：0-普通券，1-秒杀券 |
| status | uint32 | 状态：1-上架，2-下架，3-过期 |
| create_time | string | 创建时间（ISO 8601格式） |
| update_time | string | 更新时间（ISO 8601格式） |

#### 秒杀券额外字段（仅当 type=1 时返回）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| stock | int32 | 剩余库存数量 |
| begin_time | string | 秒杀开始时间（ISO 8601格式） |
| end_time | string | 秒杀结束时间（ISO 8601格式） |

### 错误响应

**HTTP状态码**: `400 Bad Request` / `401 Unauthorized` / `404 Not Found` / `500 Internal Server Error`

```json
{
  "error": "错误信息描述"
}
```

#### 常见错误码

| HTTP状态码 | 错误场景 |
|-----------|---------|
| 400 | shopId 参数无效或格式错误 |
| 401 | 未登录或 token 无效 |
| 404 | 商铺不存在 |
| 500 | 服务器内部错误 |

---

## 业务逻辑说明

### 1. 数据查询逻辑

**普通券查询**：
- 查询 `tb_voucher` 表，条件：`shop_id = :shopId AND status = 1`
- 只返回状态为"上架"的优惠券

**秒杀券关联查询**：
- 当优惠券的 `type = 1` 时，需要关联查询 `tb_seckill_voucher` 表
- 关联条件：`tb_seckill_voucher.voucher_id = tb_voucher.id`
- 返回秒杀券的额外字段：stock（库存）、begin_time（开始时间）、end_time（结束时间）

### 2. 前端处理逻辑

前端接收到数据后会进行以下处理：

```javascript
// 1. 计算折扣
let discount = (v.payValue * 10) / v.actualValue; // 例如：5折

// 2. 格式化价格显示（分转元）
let price = v.payValue / 100; // 10000分 = 100元

// 3. 判断秒杀券状态
if (v.type === 1) {
  // 判断是否已开始
  let isNotBegin = new Date(v.begin_time) > new Date();
  
  // 判断是否已结束
  let isEnd = new Date(v.end_time) < new Date();
  
  // 判断库存
  let hasStock = v.stock > 0;
}

// 4. 过滤已结束的券（前端过滤）
// 如果 end_time < 当前时间，则不显示该券
```

### 3. 数据合并说明

后端需要将普通券和秒杀券的数据合并返回：
- 对于 `type = 0` 的普通券，只返回基础字段
- 对于 `type = 1` 的秒杀券，需要 JOIN `tb_seckill_voucher` 表，返回完整信息
- 建议使用 LEFT JOIN，确保即使秒杀信息缺失也能返回基础券信息

---

## 数据库查询参考

### SQL 查询示例（伪代码）

```sql
-- 方案1：分两次查询
-- 查询基础优惠券信息
SELECT * FROM tb_voucher 
WHERE shop_id = ? AND status = 1;

-- 对于秒杀券（type=1），再查询秒杀信息
SELECT * FROM tb_seckill_voucher 
WHERE voucher_id IN (?);

-- 方案2：使用 LEFT JOIN（推荐）
SELECT 
  v.*,
  sv.stock,
  sv.begin_time,
  sv.end_time
FROM tb_voucher v
LEFT JOIN tb_seckill_voucher sv ON v.id = sv.voucher_id AND v.type = 1
WHERE v.shop_id = ? AND v.status = 1
ORDER BY v.create_time DESC;
```

### 数据库表结构

#### tb_voucher（优惠券表）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | bigint unsigned | 主键 |
| shop_id | bigint unsigned | 商铺ID |
| title | varchar(255) | 代金券标题 |
| sub_title | varchar(255) | 副标题 |
| rules | varchar(1024) | 使用规则 |
| pay_value | bigint unsigned | 支付金额（分） |
| actual_value | bigint | 抵扣金额（分） |
| type | tinyint unsigned | 0-普通券，1-秒杀券 |
| status | tinyint unsigned | 1-上架，2-下架，3-过期 |
| create_time | timestamp | 创建时间 |
| update_time | timestamp | 更新时间 |

#### tb_seckill_voucher（秒杀券表）

| 字段名 | 类型 | 说明 |
|--------|------|------|
| voucher_id | bigint unsigned | 关联的优惠券ID（主键） |
| stock | int | 库存 |
| create_time | timestamp | 创建时间 |
| begin_time | timestamp | 生效时间 |
| end_time | timestamp | 失效时间 |
| update_time | timestamp | 更新时间 |

---

## 实现要点

### 后端开发注意事项：

1. **参数验证**
   - shopId 必须是有效的正整数
   - 需要验证 token 的有效性

2. **数据关联**
   - 使用 LEFT JOIN 关联秒杀券表
   - 确保普通券不返回秒杀相关字段（stock、begin_time、end_time）
   - 秒杀券必须返回完整的秒杀信息

3. **性能优化**
   - 对 `shop_id` 和 `status` 字段建立联合索引
   - 考虑使用 Redis 缓存热门商铺的优惠券列表
   - 秒杀券库存信息建议实时从 Redis 读取

4. **数据过滤**
   - 只返回 `status = 1`（上架）的优惠券
   - 可以在后端过滤掉已过期的券（end_time < 当前时间）
   - 或者让前端处理过期判断逻辑

5. **时间格式**
   - 统一使用 ISO 8601 格式返回时间
   - 或使用 Unix 时间戳（毫秒）

---

## 测试用例

### 测试场景1: 查询普通商铺优惠券
```
请求: GET /api/voucher/list/1
期望: 返回商铺1的所有优惠券（包括普通券和秒杀券）
```

### 测试场景2: 商铺无优惠券
```
请求: GET /api/voucher/list/999
期望: 返回空数组 []
```

### 测试场景3: 无效的shopId
```
请求: GET /api/voucher/list/abc
期望: 返回 400 错误
```

### 测试场景4: 未登录访问
```
请求: GET /api/voucher/list/1 (不带token)
期望: 返回 401 错误
```

### 测试场景5: 秒杀券数据完整性
```
请求: GET /api/voucher/list/1
期望: 
- type=0 的券不包含 stock, begin_time, end_time
- type=1 的券必须包含 stock, begin_time, end_time
```

---

## 相关接口

- `GET /api/shop/:id` - 获取商铺详情
- `POST /api/voucher-order/seckill/:id` - 秒杀抢购优惠券
- `GET /api/user/me` - 获取当前登录用户信息

---

## 前端调用示例

```javascript
// 查询优惠券列表
axios.get("/voucher/list/" + shopId)
  .then(({data}) => {
    this.vouchers = data;
    
    // 前端会进行以下处理：
    // 1. 过滤已结束的券
    // 2. 计算折扣显示
    // 3. 格式化价格显示
    // 4. 判断秒杀状态（未开始/进行中/已结束/库存不足）
  })
  .catch(err => {
    console.error("查询优惠券失败", err);
  });
```

---

**文档版本**: v1.0  
**创建时间**: 2025-11-11  
**维护人员**: Backend Team