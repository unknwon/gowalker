{{define "labelmodal_en"}}
			<div id="label_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
				<form id="label_form" class="modal-form">
					<div class="modal-header">
						<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
						<h3 id="myModalLabel" style="margin: 10px">Add or remove tags.<small> <a target="_blank" href="/about#label">Learn more</a></small></h3>
					</div>
					<div class="modal-body" style="overflow-y: visible;">
						<table class="table table-condensed table-bordered">
							<tbody>
								<tr>
									<td><code>wf</code></td>
									<td>Web Framework</td>
									<td><code>orm</code></td>
									<td>Object Relation Mapping</td>
								</tr>
								<tr>
									<td><code>dbd</code></td>
									<td>Database Driver</td>
									<td><code>gui</code></td>
									<td>Graphic User Interface</td>
								</tr>
								<tr>
									<td><code>net</code></td>
									<td>Networking</td>
									<td><code>tool</code></td>
									<td>Toolkit</td>
								</tr>
							</tbody>
						</table>
						<input id="label_box" autofocus="autofocus" class="span5" autocomplete="off" type="text" placeholder="type label" data-source="[{{str2html .LabelDataSrc}}]" data-provide="typeahead">
					</div>
					<div class="modal-footer">
						<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">Close</button>
						<button type='button' class="btn" onclick="AddLabelSubmit(this)">Add</button>
						<button type='button' class="btn btn-primary" onclick="RemoveLabelSubmit(this)">Remove</button>
					</div>
				</form>
			</div>
{{end}}

{{define "labelmodal_zh"}}
			<div id="label_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
				<form id="label_form" class="modal-form">
					<div class="modal-header">
						<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
						<h3 id="myModalLabel" style="margin: 10px">增加或移除标签.<small> <a target="_blank" href="/about#tags">了解更多</a></small></h3>
					</div>
					<div class="modal-body" style="overflow-y: visible;">
						<table class="table table-condensed table-bordered">
							<tbody>
								<tr>
									<td><code>wf</code></td>
									<td>Web 框架</td>
									<td><code>orm</code></td>
									<td>对象关系映射</td>
								</tr>
								<tr>
									<td><code>dbd</code></td>
									<td>数据库驱动</td>
									<td><code>gui</code></td>
									<td>图形用户界面</td>
								</tr>
								<tr>
									<td><code>net</code></td>
									<td>计算机网络</td>
									<td><code>tool</code></td>
									<td>工具包</td>
								</tr>
							</tbody>
						</table>
						<input id="label_box" autofocus="autofocus" class="span5" autocomplete="off" type="text" placeholder="请输入标签" data-source="[{{str2html .LabelDataSrc}}]" data-provide="typeahead">
					</div>
					<div class="modal-footer">
						<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">关闭</button>
						<button type='button' class="btn" onclick="AddLabelSubmit(this)">增加
						<button type='button' class="btn btn-primary" onclick="RemoveLabelSubmit(this)">移除</button>
					</div>
				</form>
			</div>
{{end}}
