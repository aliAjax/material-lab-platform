# Bug

worker 丢失上游 context，活跃 handler 与队列状态转换在取消后继续运行。

# 触发方式

启动一个阻塞 handler，等它进入后取消 Runner context；另用已取消 context 调用队列租约，并取消 worker 的父 context。

# 错误信息

`runner did not propagate cancellation`
