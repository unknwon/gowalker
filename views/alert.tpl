{{define "alert_en"}}
{{if .IsBeta}}
<div class="alert">
    <button type="button" class="close" data-dismiss="alert">&times;</button>
    <strong>Warning!</strong> 
    This is the <span class="label label-important">Beta</span> version of new release of Go Walker, errors may occur at any time(old data will be imported in few days).
</div>
{{end}}

{{if .IsHasError}}
<div class="alert alert-error">
    <button type="button" class="close" data-dismiss="alert">&times;</button>
    <strong>ERROR</strong> 
    Your previous operation returns: {{.ErrMsg}}.
</div>
{{end}}
{{end}}

{{define "alert_zh"}}
{{if .IsBeta}}
 <div class="alert">
    <button type="button" class="close" data-dismiss="alert">&times;</button>
    <strong>注意事项</strong> 
    当前运行的 Go 步行者版本为 <span class="label label-important">测试版</span>，任何错误都有可能发生（旧版本的数据将会在几天内导入）.
</div>
{{end}}

{{if .IsHasError}}
<div class="alert alert-error">
    <button type="button" class="close" data-dismiss="alert">&times;</button>
    <strong>错误！</strong> 
    您的前次操作返回信息：{{.ErrMsg}}.
</div>
{{end}}
{{end}}