{{str2html .PkgFullIntro}}

<!-- START: Index -->
{{if .IsHasExports}}
<h3 id="_index">
	Index 
	{{if .IsCmd}}
	{{else}}
	<span class="label label-{{.DocCPLabel}} label-sm">
		Documentation complete {{.DocCP}}
	</span>
	{{end}}
</h3>

<ul class="unstyled">
	{{if .IsHasConst}}
	<li>
		<a href="#_constants">Constants</a>
	</li>
	{{end}}
	
	{{if .IsHasVar}}
	<li>
		<a href="#_variables">Variables</a>
	</li>
	{{end}}
	
	{{range .Funcs}}
	<li>
		<a href="#{{.Name}}">{{.Decl}}</a>
	</li>
	{{end}}

	{{range .Types}}
	{{$typeName := .Name}}
	<li>
		<a href="#{{.Name}}">type {{.Name}}</a>
	</li>
	
	<ul>
	{{range .Funcs}}
		<li>
			<a href="#{{.Name}}">{{.Decl}}</a>
		</li>
	{{end}}
	
	{{range .Methods}}
		<li>
			<a href="#{{$typeName}}_{{.Name}}">{{.Decl}}</a>
		</li>
	{{end}}
	</ul>
	{{end}}
</ul>

{{if .IsHasExample}}
<h3 id="_exams">Examples</h3>
<ul class="unstyled">
	{{range .Examples}}
	{{if .IsUsed}}
	<li><a href="#_ex_btn_{{.Name}}" onclick="showExample({{.Name}})">{{.Name}}</a></li>
	{{else}}
	{{template "example" .}}
	{{end}}
	{{end}}
</ul>
{{end}}

{{template "tpl/exportmodal.tpl" .}}
{{end}}
<b></b>
<!-- END: Index -->

<!-- START: Constants -->
{{if .IsHasConst}}
	<h3 id="_constants">Constants</h3>
	{{with .Consts}}
		{{range .}}
			<pre class="pre-x-scrollable pre-var">{{str2html .FmtDecl}}</pre>
			{{str2html .Doc}}
		{{end}}
	{{end}}
{{end}}
<!-- END: Constants -->

<!-- START: Variables -->
{{if .IsHasVar}}
	<h3 id="_variables">Variables</h3>
	{{with .Vars}}
		{{range .}}
			<pre class="pre-x-scrollable pre-var">{{str2html .FmtDecl}}</pre>
			{{str2html .Doc}}
		{{end}}
	{{end}}
{{end}}
<b></b>
<!-- END: Variables -->

{{$secure := .Secure}}
<!-- START: Functions -->
{{range .Funcs}}
<h3 id="{{.Name}}">
	func 
	<a target="_blank" href="http{{$secure}}://{{.URL}}">{{.Name}}</a> 
	<button class="btn btn-info btn-xs" data-toggle="collapse" data-target="#collapse_{{.Name}}" onclick="viewCode(decl|||{{.Name}});">View Code</button>
</h3>

<div class="panel panel-default">
	<div class="pre-x-scrollable">
		<pre id="decl_{{.Name}}">{{str2html .FmtDecl}}</pre>
	</div>
	<div id="collapse_{{.Name}}" class="panel-collapse collapse">
		<pre class="code">{{str2html .Code}}</pre>
	</div>
</div>

{{str2html .Doc}}

{{if isHasEleE .Examples}}
{{range .Examples}}
{{template "example" .}}
{{end}}
{{end}}
{{end}}
<b></b>
<!-- END: Functions -->

<!-- START: Types -->
{{range .Types}}
<h3 id="{{.Name}}">
	type 
	<a target="_blank" href="http{{$secure}}://{{.URL}}">{{.Name}}</a>
</h3>

<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
<br>
{{str2html .Doc}}

{{if isHasEleE .Examples}}
{{range .Examples}}
{{template "example" .}}
{{end}}
{{end}}

<!-- START: Types.Constants -->
{{with .Consts}}
{{range .}}
	<pre class="pre-x-scrollable pre-var">{{str2html .FmtDecl}}</pre>
	<p>{{str2html .Doc}}</p>
{{end}}
{{end}}
<!-- END: Types.Constants -->

<!-- START: Types.Variables -->
{{with .Vars}}
{{range .}}
	<pre class="pre-x-scrollable pre-var">{{str2html .FmtDecl}}</pre>
	<p>{{str2html .Doc}}</p>
{{end}}
{{end}}
<b></b>
<!-- END: Types.Variables -->

<!-- START: Types.Functions -->
{{range .Funcs}}
<h4 id="{{.Name}}">
	func 
	<a target="_blank" href="http{{$secure}}://{{.URL}}">{{.Name}}</a> 
	<button class="btn btn-info btn-xs" data-toggle="collapse" data-target="#collapse_{{.Name}}" onclick="viewCode(decl|||{{.Name}});">View Code</button>
</h4>

<div class="panel panel-default">
	<div class="pre-x-scrollable">
		<pre id="decl_{{.Name}}">{{str2html .FmtDecl}}</pre>
	</div>
	<div id="collapse_{{.Name}}" class="panel-collapse collapse">
		<pre class="code">{{str2html .Code}}</pre>
	</div>
</div>

{{str2html .Doc}}

{{if isHasEleE .Examples}}
{{range .Examples}}
{{template "example" .}}
{{end}}
{{end}}
{{end}}
<b></b>
<!-- END: Types.Functions -->

<!-- START: Types.Methods -->
{{range .Methods}}
<h4 id="{{.FullName}}">
	func 
	<a target="_blank" href="http{{$secure}}://{{.URL}}">{{.Name}}</a> 
	<button class="btn btn-info btn-xs" data-toggle="collapse" data-target="#collapse_{{.FullName}}" onclick="viewCode(decl|||{{.Name}});">View Code</button>
</h4>

<div class="panel panel-default">
	<div class="pre-x-scrollable">
		<pre id="decl_{{.Name}}">{{str2html .FmtDecl}}</pre>
	</div>
	<div id="collapse_{{.FullName}}" class="panel-collapse collapse">
		<pre class="code">{{str2html .Code}}</pre>
	</div>
</div>

{{str2html .Doc}}

{{if isHasEleE .Examples}}
{{range .Examples}}
{{template "example" .}}
{{end}}
{{end}}
{{end}}
<b></b>
<!-- END: Types.Methods -->
{{end}}
<b></b>
<!-- END: Types -->

{{if .IsHasFiles}}
<h3 id="_files">
	<a target="_blank" href="http{{$secure}}://{{.ViewFilePath}}">Files</a>
	{{if .IsHasHv}}
    <small class="text-info">Click any file to enable Hacker View.</small>
    {{end}}
</h3>
{{end}}

<p>
    {{if .IsHasHv}}
    {{$importPath := .ImportPath}}
	{{range .Files}}
	<a target="_blank" href="/{{$importPath}}?f={{.SrcName}}">{{.SrcName}}</a>
	{{end}}
    {{else}}
	{{range .Files}}
	<a target="_blank" href="http{{$secure}}://{{.BrowseUrl}}">{{.SrcName}}</a>
	{{end}}
    {{end}}
</p>

{{if .IsHasSubdirs}}
<h3 id="_subdirs">
	<a target="_blank" href="http{{$secure}}://{{.ViewDirPath}}">Directories</a>
</h3>

<table class="table table-hover table-striped table-condensed">
	<thead>
		<tr>
			<th>Import Path</th>
			<th>Synopsis</th>
		</tr>
	</thead>

	<tbody>
		{{range .Subdirs}}
		<tr>
			<td><a href="/{{.ImportPath}}">{{.ImportPath}}</a></td>
			<td>{{.Synopsis}}</td>
		</tr>
		{{end}}
	</tbody>
</table>
{{end}}

{{define "example"}}
<div class="panel panel-default panel-ex">
	<div class="panel-heading">
		<h4 class="panel-title">
			<a id="_ex_btn_{{.Name}}" class="panel-toggle collapsed" data-toggle="collapse" href="#_ex_{{.Name}}">Example({{.Name}})</a>
		</h4>

		<div id="_ex_{{.Name}}" class="panel-collapse collapse">
			<div class="panel-body">
				<p>Code:</p>
				<pre class="pre-x-scrollable">{{str2html .Code}}</pre>
				{{if isNotEmptyS .Output}}<p>Output:</p><pre class="pre-x-scrollable">{{.Output}}</pre>{{end}}
			</div>
		</div>
	</div>
</div>
{{end}}