# REST API v1

除公开核验、登录和刷新外，均要求 Bearer 访问令牌。写接口支持客户端设置 `X-Request-ID` 用于审计关联；生产部署应再增加持久化 `Idempotency-Key` 响应缓存。

| Method | Path | Roles | Purpose |
|---|---|---|---|
| POST | `/api/v1/auth/login` | public | 登录并获取访问/刷新 token |
| POST | `/api/v1/auth/refresh` | public | 单次轮换刷新 token |
| POST | `/api/v1/auth/logout` | signed-in | 撤销指定会话 |
| GET | `/api/v1/me` | signed-in | 当前账号和实验室配置 |
| GET/DELETE | `/api/v1/sessions[/{id}]` | signed-in | 管理自己的刷新会话 |
| GET/POST | `/api/v1/commissions` | signed-in / registrar,manager | 查询/新建委托 |
| GET/PATCH | `/api/v1/commissions/{id}` | signed-in / registrar,manager | 查看/编辑草稿 |
| POST | `/api/v1/commissions/{id}/submit` | registrar,manager | 提交并生成样品 |
| GET | `/api/v1/samples[/{id}]` | signed-in | 样品列表/父子详情 |
| POST | `/api/v1/samples/{id}/splits` | registrar,manager | 拆分子样 |
| GET/POST | `/api/v1/custody-transfers` | signed-in | 查询/发起交接 |
| POST | `/api/v1/custody-transfers/{id}/confirm` | recipient | 接收方确认 |
| GET/POST | `/api/v1/methods` | signed-in / manager | 查询/创建方法 |
| GET | `/api/v1/methods/{id}` | signed-in | 查看固定版本 |
| POST | `/api/v1/methods/{id}/validate` | signed-in | 校验公式和字段 |
| POST | `/api/v1/methods/{id}/publish` | manager | 发布不可变版本 |
| GET/POST | `/api/v1/tasks` | signed-in / registrar,manager | 查询/创建任务 |
| GET | `/api/v1/tasks/{id}` | signed-in | 任务、方法、轮次和复测 |
| POST | `/api/v1/tasks/{id}/start` | assignee | 持有子样后开始 |
| POST | `/api/v1/tasks/{id}/rounds` | executor | 写原始读数并服务端重算 |
| POST | `/api/v1/tasks/{id}/submit-review` | executor | 冻结并提交复核 |
| GET | `/api/v1/reviews[/{id}]` | signed-in | 待复核列表/详情 |
| POST | `/api/v1/reviews/{id}/decision` | reviewer,manager | 通过或退回 |
| POST | `/api/v1/certificates/issue` | manager | 签发证书 |
| GET | `/api/v1/certificates[/{id}]` | signed-in | 列表/打印详情 |
| GET | `/api/v1/public/certificates/verify?code=` | public | 最小披露公开核验 |
| GET | `/api/v1/audits` | manager | 只读审计游标流 |

## 示例

```bash
curl -sS http://localhost:18080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"registrar","password":"Registrar123!"}'
```

```bash
curl -sS http://localhost:18080/api/v1/commissions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"organization":"东海特钢","materialGrade":"Q690D","batchNumber":"B-2408","sampleDescription":"热轧板拉伸样","requirements":"GB/T 228.1 拉伸","receivedAt":"2026-08-19T01:00:00Z","totalQuantity":"10","unit":"piece"}'
```
