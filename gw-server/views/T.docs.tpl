{{str2html .PkgFullIntro}}

<!-- START: Index -->
{{if .IsHasExports}}
<h3 id="_index">
	Index 
	{{if .IsCmd}}
	{{else}}
	<span class="label label-{{.DocCPLabel}}">
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
	
	{{range .Funcs}}
	<ul>
		<li>
			<a href="#{{.Name}}">{{.Decl}}</a>
		</li>
	</ul>
	{{end}}
	
	{{range .Methods}}
	<ul>
		<li>
			<a href="#{{$typeName}}_{{.Name}}">{{.Decl}}</a>
		</li>
	</ul>
	{{end}}
	{{end}}
</ul>

{{if .IsHasExams}}
<h3 id="_exams">Examples</h3>
<ul class="unstyled">
	{{range .Exams}}
	{{if .IsUsed}}
	<li><a href="#_ex_btn_{{.Name}}" onclick="showExample({{.Name}})">{{.Name}}</a></li>
	{{else}}
	{{template "example_en" .}}
	{{end}}
	{{end}}
</ul>
{{end}}

{{template "exportmodal_en" .}}
{{end}}
<b></b>
<!-- END: Index -->

<!-- START: Constants -->
{{if .IsHasConst}}
	<h3 id="_constants">Constants</h3>
	{{with .Consts}}
		{{range .}}
			<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
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
			<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
			{{str2html .Doc}}
		{{end}}
	{{end}}
{{end}}
<b></b>
<!-- END: Variables -->

<!-- START: Functions -->
{{range .Funcs}}
<h3 id="{{.Name}}">
	func 
	<a target="_blank" href="{{.URL}}">{{.Name}}</a> 
	<small>
		<a class="accordion-toggle" data-toggle="collapse" href="#collapse_{{.Name}}">View Code</a>
	</small>
</h3>

<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
{{str2html .Doc}}
<div class="accordion">
		<div id="collapse_{{.Name}}" class="accordion-body collapse">
		<pre class="accordion-inner">{{str2html .Code}}</pre>
	</div>
</div>

{{if .IsHasExam}}
{{range .Exams}}
{{template "example_en" .}}
{{end}}
{{end}}
{{end}}
<b></b>
<!-- END: Functions -->

<!-- START: Types -->
{{range .Types}}
<h3 id="{{.Name}}">
	type 
	<a target="_blank" href="{{.URL}}">{{.Name}}</a>
</h3>

<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
{{str2html .Doc}}

{{if .IsHasExam}}
{{range .Exams}}
{{template "example_en" .}}
{{end}}
{{end}}

<!-- START: Types.Constants -->
{{with .Consts}}
{{range .}}
	<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
	<p>{{str2html .Doc}}</p>
{{end}}
{{end}}
<!-- END: Types.Constants -->

<!-- START: Types.Variables -->
{{with .Vars}}
{{range .}}
	<pre class="pre-x-scrollable">{{str2html .FmtDecl}}</pre>
	<p>{{str2html .Doc}}</p>
{{end}}
{{end}}
<b></b>
<!-- END: Types.Variables -->

<!-- START: Types.Functions -->
{{range .Funcs}}
<h4 id="{{.Name}}">
	func 
	<a target="_blank" href="{{.URL}}">{{.Name}}</a>
	<small>
		<a class="accordion-toggle" data-toggle="collapse" data-parent="#accordion" href="#collapse_{{.Name}}">View Code</a>
	</small>
</h4>

<pre>{{str2html .FmtDecl}}</pre>
{{str2html .Doc}}
<div class="accordion">
		<div id="collapse_{{.Name}}" class="accordion-body collapse">
		<pre class="accordion-inner">{{str2html .Code}}</pre>
	</div>
</div>

{{if .IsHasExam}}
{{range .Exams}}
{{template "example_en" .}}
{{end}}
{{end}}
{{end}}
<b></b>
<!-- END: Types.Functions -->

<!-- START: Types.Methods -->
{{range .Methods}}
<h4 id="{{.FullName}}">
	func 
	<a target="_blank" href="{{.URL}}">{{.Name}}</a> 
	<small>
		<a class="accordion-toggle" data-toggle="collapse" data-parent="#accordion" href="#collapse_{{.FullName}}">View Code</a>
	</small>
</h4>

<pre>{{str2html .FmtDecl}}</pre>
{{str2html .Doc}}
	<div class="accordion">
		<div id="collapse_{{.FullName}}" class="accordion-body collapse">
		<pre class="accordion-inner">{{str2html .Code}}</pre>
	</div>
	</div>

{{if .IsHasExam}}
{{range .Exams}}
{{template "example_en" .}}
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
	<a target="_blank" href="http://{{.ProPath}}">Files</a>
</h3>
{{end}}

<p>
	{{$proPath := .ProPath}}
	{{$pkgTag := .PkgTag}}
	{{range .Files}}
	<a target="_blank" href="http://{{$proPath}}/{{.}}{{$pkgTag}}">{{.}}</a>
	{{end}}
</p>

{{if .IsHasSubdirs}}
<h3 id="_subdirs">
	<a target="_blank" href="http://{{.ProPath}}">Directories</a>
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
			<td><a href="/{{.Path}}">{{.Path}}</a></td>
			<td>{{.Synopsis}}</td>
		</tr>
		{{end}}
	</tbody>
</table>
{{end}}