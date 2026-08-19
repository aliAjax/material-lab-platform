# Bug

附件写入与清理同时失败时主错误被覆盖，审计校验错误也无法通过 errors.Is 分类。

# 触发方式

使用同时返回写入错误和删除错误的 Storage，并向审计追加校验传入缺失字段和倒退时间。

# 错误信息

`missing write failure: cleanup unavailable`
