<div id="example_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
	<form id="example_form" class="modal-form" action="/examples">
		<div class="modal-header">
			<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
			<h3 id="myModalLabel" style="margin: 10px">Add examples.</h3>
		</div>
		<input type="hidden" name="q" value="{{.ImportPath}}">	
		<div class="modal-body" style="overflow-y: visible;">
			<input id="example_box" autofocus="autofocus" class="span5" type="text" name="gist" placeholder="type Gist address">
		</div>

		<div class="modal-footer">
			<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">Close</button>
			<button class="btn btn-primary" type="submit">Go!</button>
		</div>
	</form>
</div>

{{define "exammodal_zh"}}
<div id="example_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
	<form id="example_form" class="modal-form" action="/examples">
		<div class="modal-header">
			<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
			<h3 id="myModalLabel" style="margin: 10px">增加示例.</h3>
		</div>
		<input type="hidden" name="q" value="{{.ImportPath}}">
		<div class="modal-body" style="overflow-y: visible;">
			<input id="example_box" autofocus="autofocus" class="span5" type="text" name="gist" placeholder="请输入 Gist 地址">
		</div>

		<div class="modal-footer">
			<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">关闭</button>
			<button class="btn btn-primary" type="submit">增加</button>
		</div>
	</form>
</div>
{{end}}