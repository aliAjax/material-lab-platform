# 材料检测实验室样品流转与双人复核平台

面向金属、涂层和塑料材料检测实验室的全过程工作台。项目覆盖送检登记、子样拆分、双人交接、版本化检测方法、精确试验计算、异常复测、技术复核、证书签发和公开真伪核验。它不是通用 RBAC、库存、CRM、预约或报表系统。

## 技术架构

- 后端：Go 1.25、Chi、JWT、bcrypt、shopspring/decimal、pgx。
- 前端：React 18、TypeScript、Vite、TanStack Query、React Hook Form、Zod、Lucide。
- 数据：提供完整 PostgreSQL 17 迁移定义和默认运行时 repository。设置 `DATABASE_URL` 后服务会连接 PostgreSQL，并将领域聚合持久化到关系型 schema 的受控存储表；未设置该变量时仅用于本地无依赖演示的线程安全内存适配器，重启会清空数据。
- 文件：受控本地文件存储，接口预留为 `attachment.Storage`，可替换对象存储。
- 异步：数据库任务表的 schema，以及带租约、退避重试和幂等完成语义的 Go worker。

领域层不依赖 HTTP、数据库或前端类型。应用层编排事务，`ports.Repository` 定义持久化边界，HTTP、内存和 PostgreSQL 连接池位于适配器层。重要写操作追加审计事件；统一错误体始终带请求 ID。

```text
cmd/server                 HTTP 服务、演示初始化
cmd/worker                 后台任务 worker
internal/domain            领域实体、状态机、公式 AST
internal/application       业务用例与事务编排
internal/ports             存储端口
internal/adapters/http     REST、认证中间件、错误映射
internal/adapters/memory   可运行演示存储
internal/adapters/postgres PostgreSQL 连接池适配器
internal/auth              登录、JWT、刷新会话、限流
internal/{sample,custody,method,execution,review,certificate}
                            业务策略与展示支持
internal/{attachment,audit,notification,health,observability,worker}
                            基础能力
migrations                 PostgreSQL 迁移
frontend/src/features      按业务功能拆分的页面
deployments                Docker Compose
docs                       API 与状态流转说明
```

## 快速启动

要求 Go 1.25+ 和 Node.js 22+。

```bash
go mod download
go run ./cmd/server
```

另开终端启动前端：

```bash
cd frontend
npm install
npm run dev
```

访问 `http://localhost:15173`。Vite 会把 `/api` 和健康检查代理到 `http://localhost:18080`。

### 演示账号

| 职责 | 用户名 | 密码 |
|---|---|---|
| 实验室负责人 | `manager` | `Manager123!` |
| 登记员 | `registrar` | `Registrar123!` |
| 试验员 | `tester` | `Tester123!` |
| 复核员 | `reviewer` | `Reviewer123!` |

这些账号只用于本地演示。生产环境必须通过初始化命令创建负责人、关闭演示种子并替换 `JWT_SECRET`。

### Docker Compose

先构建前端静态文件，再启动服务：

```bash
cd frontend && npm install && npm run build && cd ..
docker compose -f deployments/docker-compose.yml up --build
```

API 为 `http://localhost:18080`，前端为 `http://localhost:15173`，PostgreSQL 宿主端口为 `15432`。迁移在新数据卷首次创建时执行。

关闭并清理运行容器（保留数据卷）：

```bash
docker compose -f deployments/docker-compose.yml down
```

## 完整操作路径

1. 使用登记员登录，进入“新建委托”，填写来源、牌号、批号、描述、要求、接收时间，保存草稿并提交，得到唯一 `SMP-YYYYMMDD-NNNNN` 样品号。
2. 在样品详情拆分子样。子样用途、数量、单位和位置必填；发生损耗必须解释，总分配不得超过父样数量。
3. 发起交接给试验员。切换试验员账号确认接收；发送方不能替对方确认，未确认记录不算完成。
4. 负责人可以查看内置的拉伸试验方法，也可新建、校验并发布方法版本。发布版本不可编辑。
5. 登记员或负责人创建任务并指派试验员。试验员只有确认持有子样后才能开始，根据方法字段录入原始数据。后端使用十进制 AST 重算公式并按 HALF_UP 精度舍入，不接受前端计算结果。
6. 试验员提交技术复核后数据冻结。使用不同的复核员账号核对读数和公式，填写意见后通过或退回；执行者不能自审。
7. 全部任务通过后，负责人签发不可覆盖的证书。响应只在签发时返回一次高熵核验码。
8. 退出登录，在 `/verify?code=...` 公开验证。公开响应只含有效性、证书号、状态、签发时间和摘要，不暴露客户、样品或同一委托信息。

## 状态机

- 委托：`draft -> submitted -> completed`
- 样品：`registered -> split -> in_testing -> sealed -> disposed`
- 交接：`pending -> confirmed`；纠错通过关联原记录的新 `reversal`，不修改已确认事实。
- 方法：`draft -> published -> archived`；发布后新变化必须创建新版本。
- 任务：`pending -> in_progress -> review -> approved`，或 `review -> rejected -> in_progress` 产生新轮次。
- 证书：`issued -> voided`；更正证书通过 `supersedes_id` 关联原证书。

关键迁移均在应用事务内读取当前状态再更新。生产 PostgreSQL schema 使用唯一约束、检查约束、外键、游标索引和不可更新/删除的审计触发器作为第二道保护。

## 公式与计量

受限表达式解析器仅接受字段名、十进制字面量、`+ - * /`、括号以及 `min`、`max`、`avg`。解析发布时拒绝未知变量和非法 token；运行时拒绝除零。所有读数和结果使用任意精度十进制值，原始读数完整保存，最终值按方法配置的小数位执行 `shopspring/decimal.Round`（HALF_UP）舍入。

## API

所有业务 API 位于 `/api/v1`。认证使用 `Authorization: Bearer <accessToken>`，访问令牌默认 15 分钟；刷新令牌单次轮换且可撤销。接口清单与例子见 [docs/api.md](docs/api.md)。

错误结构：

```json
{"error":{"code":"validation","message":"validation failed: field required","requestId":"...","fields":{}}}
```

所有时间通过 RFC 3339 传输，服务端和 PostgreSQL 保存 UTC，前端根据实验室时区显示。持续增长列表响应为 `{items,nextCursor}`，不把 offset 作为协议。

## 测试与质量检查

```bash
gofmt -w $(find . -name '*.go' -not -path './frontend/*')
go vet ./...
go test ./...
cd frontend
npm run typecheck
npm test
npm run build
```

测试重点覆盖精确表达式、除零与未知变量、拆分约束、双人交接、自审禁止、证书不可重复作废、附件 MIME/路径/越权和 worker 重复消费。HTTP 主链路可用 `docs/e2e.sh` 验证。

按题目口径统计非测试 Go 有效代码（排除空行、纯注释、`*_test.go`、vendor）：

```bash
find . -name '*.go' ! -name '*_test.go' ! -path './vendor/*' -print0 \
  | xargs -0 awk 'NF && $1 !~ /^\/\//' | wc -l
```

## 安全边界

- bcrypt 密码哈希；登录按来源地址做失败限流；刷新 token 仅保存摘要并轮换。
- 请求体限制、超时、panic 恢复、请求 ID、安全响应头和角色职责校验。
- 附件使用服务端随机键，校验真实 MIME、扩展名、大小及访问者与业务对象关系。
- 审计摘要对 password、token 和 content 字段脱敏；数据库禁止更新或删除审计事件。
- 公开核验码为 192-bit 随机值，错误响应不区分不存在和无效状态。

## 当前实现边界

不设置 `DATABASE_URL` 时，服务使用内存 repository，便于没有 PostgreSQL 的开发者直接运行；数据会随进程退出清空。设置该变量或通过 Docker Compose 启动时，运行时 repository 会在 PostgreSQL 中恢复和保存样品、方法、任务、证书和审计聚合。PDF 未实现，证书使用可打印 HTML，避免不可复现的中文字体依赖。
