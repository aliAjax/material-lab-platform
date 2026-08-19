# Material Lab Frontend

工业材料检测实验室“样品流转与双人复核平台”的 React + TypeScript 前端。界面采用桌面优先、平板兼容的紧凑工作区布局，所有业务记录均来自后端 API，不包含浏览器内写死的演示数据。

## 技术栈与目录

- React、TypeScript、Vite
- React Router：业务路由与 URL 筛选状态
- TanStack Query：服务端状态、加载/错误/刷新
- React Hook Form + Zod：表单校验
- Lucide React：操作图标

源码按 `app`、`api`、`components`、`features`、`hooks`、`state`、`styles`、`types` 组织。各业务功能位于 `src/features`，API 契约集中在 `src/api/resources.ts`。

## 本地启动

需要 Node.js 20.19+ 或 22.12+。后端默认和前端部署在同一来源，并在 `/api/v1` 提供 API：

```bash
npm install
npm run dev
```

前后端分开启动时设置 API 地址：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

开发服务器默认监听 `http://localhost:5174`。生产部署应把 `/api/v1` 反向代理到 Go 服务，并让 SPA 未匹配路径回退到 `index.html`。

## 检查与构建

```bash
npm run typecheck
npm test
npm run build
```

生产文件输出到 `dist/`。访问令牌仅保存在 `sessionStorage`，刷新令牌由后端通过 HttpOnly Cookie 管理；401 会触发一次刷新，刷新失败则回到登录状态。所有写请求自动携带请求 ID 和幂等键。

## 主要操作路径

1. 登录后从待办工作台进入各角色队列。
2. 登记员在“样品与委托”中新建草稿并正式提交，再进入样品详情拆分子样、打印标签。
3. 在“待确认交接”发起领取或转交，由对方账号确认后才能开始试验。
4. 试验员进入任务，根据方法版本动态生成的字段录入读数；公式预览只用于核对，最终值由服务端重算。
5. 异常时提交复测申请；正常任务冻结后进入双人复核。复核页展示读数、结果及修订差异。
6. 全部必要任务复核通过后，负责人在证书页签发。公开用户访问 `/verify`，仅凭高熵校验码查询最小必要信息。
7. 负责人在“方法版本”维护字段和公式，校验通过后发布；发布版本仅可查看。个人页面可以撤销其他刷新会话。

危险动作均有确认对话框；编辑表单在刷新或关闭页面时提示未保存内容。日期由浏览器以实验室配置时区显示，最终业务时间以服务端响应为准。
