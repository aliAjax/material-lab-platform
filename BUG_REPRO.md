# Bug

结构化日志只脱敏顶层属性，嵌套 token 泄露；通知所有权错误未保留 forbidden 错误链。

# 触发方式

记录包含嵌套 authorization/refreshToken 的 slog group，并对别人的通知执行已读操作后检查 errors.Is/errors.As。

# 错误信息

`sensitive nested value leaked`
