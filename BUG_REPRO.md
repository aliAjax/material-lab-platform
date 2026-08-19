# Bug

方法定义的多项校验被转成纯字符串，validation sentinel 和字段上下文无法通过 errors.Is/errors.As 获取。

# 触发方式

提交同时包含非法字段名、空标签和未知公式变量的方法定义，再检查聚合错误链。

# 错误信息

`validation sentinel lost`
