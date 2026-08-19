# Bug

证书渲染未先验证并原子输出，作废失败会先修改状态，nil 证书还会 panic。

# 触发方式

向空证书和限长失败 writer 渲染，再以空作废理由调用状态转换，并测试 nil receiver。

# 错误信息

`validation error = <nil>` 或 nil pointer panic。
