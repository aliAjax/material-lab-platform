# Bug

nil 或 panic 的健康检查未被隔离，服务启动检查会崩溃；无数据库模式还注册了 nil 数据库检查。

# 触发方式

运行包含 nil、panic 和正常检查的聚合测试，或在 database 为 nil 时构造启动检查集合。

# 错误信息

`panic: runtime error: invalid memory address or nil pointer dereference`
