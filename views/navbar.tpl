{{define "navbar_en"}}
			<div class="navbar" style="margin-top: 20px">
				<div class="navbar-inner">
					<a class="brand" href="/">Go Walker</a>
					<ul class="nav">
						<li><a href="/">Home</a></li>
						<li><a href="/index">Index</a></li>
						<li><a href="/tags">Tags</a></li>
						<li><a href="/about">About</a></li>
					</ul>
					<form class="navbar-search pull-right" action="/">
						<input class="search-query" type="text" placeholder="Search" autofocus="autofocus"  name="q"></input>
					</form>
				</div>
			</div>
{{end}}

{{define "navbar_zh"}}
			<div class="navbar" style="margin-top: 20px">
				<div class="navbar-inner">
					<a class="brand" href="/">Go 步行者</a>
					<ul class="nav">
						<li><a href="/">首页</a></li>
						<li><a href="/index">索引</a></li>
						<li><a href="/tags">标签</a></li>
						<li><a href="/about">关于</a></li>
					</ul>
					<form class="navbar-search pull-right" action="/">
						<!--input type="hidden" name="lang" value="{{.Lang}}" /-->
						<input class="search-query" type="text" placeholder="搜索项目" autofocus="autofocus" name="q"></input>
					</form>
				</div>
			</div>
{{end}}

{{define "navbar_ja"}}
			<div class="navbar" style="margin-top: 20px">
				<div class="navbar-inner">
					<a class="brand" href="/">Go Walker</a>
					<ul class="nav">
						<li><a href="/">Home</a></li>
						<li><a href="/index">Index</a></li>
						<li><a href="/tags">Tags</a></li>
						<li><a href="/about">About</a></li>
					</ul>
					<form class="navbar-search pull-right" action="/">
						<input class="search-query" type="text" placeholder="Search" autofocus="autofocus"  name="q"></input>
					</form>
				</div>
			</div>
{{end}}