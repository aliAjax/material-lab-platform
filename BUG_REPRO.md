# Bug

交接确认和撤销的检查与状态写入没有原子化，产生 data race 和多个成功终态。

# 触发方式

16 个 goroutine 在同步屏障后对同一交接记录交替执行确认和撤销，并开启 race detector。

# 错误信息

`WARNING: DATA RACE`，同时出现 `successful terminal transitions = 8`。
