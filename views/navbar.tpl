{{define "navbar_en"}}
		<div id="fixed_top" class="navbar navbar-fixed-top">
			<div class="navbar-inner">
				<div id="navbar" class="container">
					<a class="brand" href="/">Go Walker</a>
					<ul class="nav">
						<li {{if .IsHome}}class="active"{{end}}><a href="/">Home</a></li>
						<li {{if .IsIndex}}class="active"{{end}}><a href="/index">Index</a></li>
						<li {{if .IsLabels}}class="active"{{end}}><a href="/labels">Labels</a></li>
						<li {{if .IsExamples}}class="active"{{end}}><a href="/examples">Examples</a></li>
						<li {{if .IsAbout}}class="active"{{end}}><a href="/about">About</a></li>
					</ul>
					<form id="top_search_form" class="navbar-search pull-right" action="/">
						<input id="navbar_search_box" class="search-query" type="text" placeholder="Search" name="q"></input>
					</form>
				</div>
			</div>
		</div>
{{end}}

{{define "navbar_zh"}}
		<div id="fixed_top" class="navbar navbar-fixed-top">
			<div class="navbar-inner">
				<div id="navbar">
					<a class="brand" href="/">Go 步行者</a>
					<ul class="nav">
						<li {{if .IsHome}}class="active"{{end}}><a href="/">首页</a></li>
						<li {{if .IsIndex}}class="active"{{end}}><a href="/index">索引</a></li>
						<li {{if .IsLabels}}class="active"{{end}}><a href="/labels">标签</a></li>
						<li {{if .IsExamples}}class="active"{{end}}><a href="/examples">示例</a></li>
						<li {{if .IsAbout}}class="active"{{end}}><a href="/about">关于</a></li>
					</ul>
					<form id="top_search_form" class="navbar-search pull-right" action="/">
						<input id="navbar_search_box" class="search-query" type="text" placeholder="搜索文档" name="q"></input>
					</form>
				</div>
			</div>
		</div>
{{end}}