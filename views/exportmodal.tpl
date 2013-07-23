{{define "exportmodal_en"}}
<div id="search_exports" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
	<form id="search_form" class="modal-form">
		<div class="modal-header">
			<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
			<h3 id="myModalLabel" style="margin: 10px">Search and Go to Exports.</h3>
		</div>

		<div class="modal-body" style="overflow-y: visible;">
			<input id="search_export_box" autofocus="autofocus" autocomplete="off" class="span5" type="text" placeholder="type or function name"  data-source="[{{str2html .ExportDataSrc}}]" data-provide="typeahead">
		</div>

		<div class="modal-footer">
			<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">Close</button>
			<button class="btn btn-primary" type="submit">Go!</button>
		</div>
	</form>
</div>
{{end}}

{{define "exportmodal_zh"}}
<div id="search_exports" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
	<form id="search_form" class="modal-form">
		<div class="modal-header">
			<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
			<h3 id="myModalLabel" style="margin: 10px">搜索导出对象.</h3>
		</div>

		<div class="modal-body" style="overflow-y: visible;">
			<input id="search_export_box" autofocus="autofocus" autocomplete="off" class="span5" type="text" placeholder="请输入类型或函数名"  data-source="[{{str2html .ExportDataSrc}}]" data-provide="typeahead">
		</div>

		<div class="modal-footer">
			<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">关闭</button>
			<button class="btn btn-primary" type="submit">跳转</button>
		</div>
	</form>
</div>
{{end}}