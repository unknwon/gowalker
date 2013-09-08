{{define "navbar_en"}}
<noscript>Please enable JavaScript in your browser!</noscript>
<div id="navbar_frame" class="navbar navbar-default navbar-fixed-top">
	<div class="container" style="padding-left: 30px;">
		<a class="navbar-brand" href="/">
			Go Walker
			<sub class="version">{{.AppVer}}</sub>
		</a>
		<ul class="nav navbar-nav">
			<li {{if .IsHome}}class="active"{{end}}>
				<a href="/">Home</a>
			</li>
			
			<li {{if .IsIndex}}class="active"{{end}}>
				<a href="/index">Index</a>
			</li>
			
			<li {{if .IsLabels}}class="active"{{end}}>
				<a href="/label">Label</a>
			</li>
			
			<li {{if .IsFuncs}}class="active"{{end}}>
				<a href="/function">Function</a>
			</li>
			
			<li {{if .IsExamples}}class="active"{{end}}>
				<a href="/example">Example</a>
			</li>

			<li class="dropdown">
				<a class="dropdown-toggle" data-toggle="dropdown" href="#">
					More.
					<b class="caret"></b>
				</a>
				<ul class="dropdown-menu">
					<li><a href="http://ctw.gowalker.org">Cross the Wall</a></li>
					<li><a href="http://video.gowalker.org">Online videos</a></li>
					<li><a href="http://res.gowalker.org">Resources</a></li>
					<li><a href="http://api.gowalker.org">API service</a></li>
					<li class="divider"></li>
					<li><a href="http://golangnews.com">Golang News</a></li>
					<li><a href="http://gotalks.io">Go Talks</a></li>
				</ul>
			</li>
			
			<li {{if .IsAbout}}class="active"{{end}}>
				<a href="/about">About</a>
			</li>
		</ul>
		<div class="navbar-right">
			<form  class="navbar-form form-inline" action="/">
				<div class="form-group">
					<input type="text" class="form-control" placeholder="Search" name="q">
				</div>
				{{if .IsLogin}}
				<button class="btn" onclick="return redirToLogin('?op=exit');">Log out</button>
				{{else}}
				<button class="btn" onclick="return redirToLogin('');">Log in</button>
				{{end}}
				{{template "navbar_js"}}
			</form>
		</div>
	</div>
</div>
{{end}}

{{define "navbar_zh"}}
<noscript>请启用您浏览器的 JavaScript 选项！</noscript>
<div id="navbar_frame" class="navbar navbar-default navbar-fixed-top">
	<div class="container" style="padding-left: 30px;">
		<a class="navbar-brand" href="/">
			Go Walker
			<sub class="version">{{.AppVer}}</sub>
		</a>
		<ul class="nav navbar-nav">
			<li {{if .IsHome}}class="active"{{end}}>
				<a href="/">首页</a>
			</li>
			
			<li {{if .IsIndex}}class="active"{{end}}>
				<a href="/index">索引</a>
			</li>
			
			<li {{if .IsLabels}}class="active"{{end}}>
				<a href="/label">标签</a>
			</li>
			
			<li {{if .IsFuncs}}class="active"{{end}}>
				<a href="/function">函数</a>
			</li>
			
			<li {{if .IsExamples}}class="active"{{end}}>
				<a href="/example">示例</a>
			</li>

			<li class="dropdown">
				<a class="dropdown-toggle" data-toggle="dropdown" href="#">
					更多.
					<b class="caret"></b>
				</a>
				<ul class="dropdown-menu">
					<li><a href="http://ctw.gowalker.org">仓库下载</a></li>
					<li><a href="http://video.gowalker.org">在线视频</a></li>
					<li><a href="http://res.gowalker.org">文档资源</a></li>
					<li><a href="http://api.gowalker.org">API 服务</a></li>
					<li class="divider"></li>
					<li><a href="http://golangnews.com">Golang News</a></li>
					<li><a href="http://gotalks.io">Go Talks</a></li>
				</ul>
			</li>
			
			<li {{if .IsAbout}}class="active"{{end}}>
				<a href="/about">关于</a>
			</li>
		</ul>
		<div class="navbar-right">
			<form  class="navbar-form form-inline" action="/">
				<div class="form-group">
					<input type="text" class="form-control" placeholder="搜索" name="q">
				</div>
				{{if .IsLogin}}
				<button class="btn btn-default btn-sm" onclick="return redirToLogin('?op=exit');">退出</button>
				{{else}}
				<button class="btn btn-default btn-sm" onclick="return redirToLogin('');">登录</button>
				{{end}}
				{{template "navbar_js"}}
			</form>
		</div>
	</div>
</div>
{{end}}

{{define "navbar_js"}}
<script>
	// if (document.body.clientWidth <= 1000) {
 //        window.location.href = "http://m.gowalker.org";
 //    }
	function redirToLogin(params) {
		window.location.href = "/login" + params
		return false;
	}
</script>
{{end}}