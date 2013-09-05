{{define "navbar_en"}}
<div id="navbar_frame" class="navbar navbar-fixed-top">
	<div class="navbar-inner">
		<div id="navbar" class="container">
			<a class="brand" href="/">
				Go Walker
				<sub class="version">{{.AppVer}}</sub>
			</a>
			<ul class="nav">
				<li {{if .IsHome}}class="active"{{end}}>
					<a href="/">Home</a>
				</li>
				
				<li {{if .IsIndex}}class="active"{{end}}>
					<a href="/index">Index</a>
				</li>
				
				<li {{if .IsLabels}}class="active"{{end}}>
					<a href="/labels">Labels</a>
				</li>
				
				<li {{if .IsFuncs}}class="active"{{end}}>
					<a href="/funcs">Functions</a>
				</li>
				
				<li {{if .IsExamples}}class="active"{{end}}>
					<a href="/examples">Examples</a>
				</li>

				<li>
					<div class="btn-group">
						<a class="btn dropdown-toggle" data-toggle="dropdown" href="#">
							More.
							<span class="caret"></span>
						</a>
						<ul class="dropdown-menu">
							<li><a href="http://golangnews.com">Golang News</a></li>
							<li><a href="http://video.gowalker.org">Online videos</a></li>
							<li><a href="http://res.gowalker.org">Resources</a></li>
							<li><a href="http://ctw.gowalker.org">Cross the Wall</a></li>
							<li><a href="http://api.gowalker.org">API service</a></li>
							<li><a href="http://gotalks.io">Go Talks</a></li>
						</ul>
					</div>
				</li>
				
				<li {{if .IsAbout}}class="active"{{end}}>
					<a href="/about">About</a>
				</li>
			</ul>
			<div id="top_right_box" class="pull-right">
				<form class="navbar-search" action="/">
					<input id="navbar_search_box" class="search-query" type="text" placeholder="Search" name="q">
				</form>
				<ul class="nav" style="padding-left: 10px;">
					{{if .IsLogin}}
					<li><a class="btn btn-small" href="/login?op=exit">Log out</a></li>
					{{else}}
					<li><a class="btn btn-small" href="/login">Log in</a></li>
					{{end}}
				</ul>
			</div>
		</div>
	</div>
</div>
{{end}}

{{define "navbar_zh"}}
<div id="navbar_frame" class="navbar navbar-fixed-top">
	<div class="navbar-inner">
		<div id="navbar" class="container">
			<a class="brand" href="/">
				Go Walker
				<sub class="version">{{.AppVer}}</sub>
			</a>
			<ul class="nav">
				<li {{if .IsHome}}class="active"{{end}}>
					<a href="/">首页</a>
				</li>
				
				<li {{if .IsIndex}}class="active"{{end}}>
					<a href="/index">索引</a>
				</li>
				
				<li {{if .IsLabels}}class="active"{{end}}>
					<a href="/labels">标签</a>
				</li>
				
				<li {{if .IsFuncs}}class="active"{{end}}>
					<a href="/funcs">函数</a>
				</li>
				
				<li {{if .IsExamples}}class="active"{{end}}>
					<a href="/examples">示例</a>
				</li>

				<li>
					<div class="btn-group">
						<a class="btn dropdown-toggle" data-toggle="dropdown" href="#">
							更多.
							<span class="caret"></span>
						</a>
						<ul class="dropdown-menu">
							<li><a href="http://golangnews.com">Golang News</a></li>
							<li><a href="http://video.gowalker.org">在线视频</a></li>
							<li><a href="http://res.gowalker.org">文档资源</a></li>
							<li><a href="http://ctw.gowalker.org">仓库下载</a></li>
							<li><a href="http://api.gowalker.org">API 服务</a></li>
							<li><a href="http://gotalks.io">Go Talks</a></li>
						</ul>
					</div>
				</li>
				
				<li {{if .IsAbout}}class="active"{{end}}>
					<a href="/about">关于</a>
				</li>
			</ul>
			<div id="top_right_box" class="pull-right">
				<form class="navbar-search" action="/">
					<input id="navbar_search_box" class="search-query" type="text" placeholder="搜索" name="q">
				</form>
				<ul class="nav" style="padding-left: 10px;">
					{{if .IsLogin}}
					<li><a class="btn btn-small" href="/login?op=exit">退出</a></li>
					{{else}}
					<li><a class="btn btn-small" href="/login">登录</a></li>
					{{end}}
				</ul>
			</div>
		</div>
	</div>
</div>
{{end}}