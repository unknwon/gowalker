{{define "tagmodal_en"}}
			<div id="tags_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
				<form id="tags_form" class="modal-form">
					<div class="modal-header">
						<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
						<h3 id="myModalLabel" style="margin: 10px">Add or remove tags.<small> <a target="_blank" href="/about#tags">Learn more</a></small></h3>
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
						<input id="tags_box" autofocus="autofocus" class="span5" type="text" placeholder="type tags"></input>
					</div>
					<div class="modal-footer">
						<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">Close</button>
						<button type='button' class="btn" onclick="AddTagSubmit(this)">Add</input>
						<button type='button' class="btn btn-primary" onclick="RemoveTagSubmit(this)">Remove</button>
					</div>
				</form>
			</div>
{{end}}

{{define "tagmodal_zh"}}
			<div id="tags_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
				<form id="tags_form" class="modal-form">
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
						<input id="tags_box" autofocus="autofocus" class="span5" type="text" placeholder="请输入标签"></input>
					</div>
					<div class="modal-footer">
						<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">关闭</button>
						<button type='button' class="btn" onclick="AddTagSubmit(this)">增加</input>
						<button type='button' class="btn btn-primary" onclick="RemoveTagSubmit(this)">移除</button>
					</div>
				</form>
			</div>
{{end}}

{{define "tagmodal_ja"}}
			<div id="tags_modal" class="modal hide fade" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
				<form id="tags_form" class="modal-form">
					<div class="modal-header">
						<button type="button" class="close" data-dismiss="modal" aria-hidden="true">×</button>
						<h3 id="myModalLabel" style="margin: 10px">Add or remove tags.<small> <a target="_blank" href="/about#tags">Learn more</a></small></h3>
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
						<input id="tags_box" autofocus="autofocus" class="span5" type="text" placeholder="type tags"></input>
					</div>
					<div class="modal-footer">
						<button type='button' class="btn" data-dismiss="modal" aria-hidden="true">Close</button>
						<button type='button' class="btn" onclick="AddTagSubmit(this)">Add</input>
						<button type='button' class="btn btn-primary" onclick="RemoveTagSubmit(this)">Remove</button>
					</div>
				</form>
			</div>
{{end}}