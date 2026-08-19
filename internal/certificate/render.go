package certificate

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"material-lab-platform/internal/domain"
)

var printTemplate = template.Must(template.New("certificate").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>{{.Number}}</title>
<style>body{font-family:sans-serif;max-width:900px;margin:40px auto;color:#17202a}h1{text-align:center}table{border-collapse:collapse;width:100%}td,th{border:1px solid #888;padding:8px}@media print{button{display:none}}</style>
</head><body><h1>检验证书</h1><table><tr><th>证书编号</th><td>{{.Number}}</td></tr><tr><th>结论</th><td>{{.Decision}}</td></tr><tr><th>签发时间</th><td>{{.IssuedAt}}</td></tr><tr><th>SHA-256摘要</th><td>{{.Digest}}</td></tr></table><button onclick="print()">打印</button></body></html>`))

func RenderHTML(destination io.Writer, value domain.Certificate) error {
	var output bytes.Buffer
	if err := printTemplate.Execute(&output, value); err != nil {
		return err
	}
	if destination == nil {
		return nil
	}
	_, err := io.Copy(destination, &output)
	if err != nil {
		return fmt.Errorf("certificate output failed: %v", err)
	}
	return nil
}
