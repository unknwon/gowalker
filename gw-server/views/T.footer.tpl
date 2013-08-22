{{define "footer_en"}}
<div id='_keyshortcut' class="modal hide fade" tabindex="-1" role="dialog" aria-hidden="false">
  <div class="modal-header">
      <h3>Control panel</h3>
  </div>

  <div class="modal-body">
      <table>
          <tbody>
	      <tr><td align="right"><b>?</b></td><td> : This menu</td></tr>
	      <tr><td align="right"><b>/</b></td><td> : Search site</td></tr>
	      <tr><td align="right"><b>.</b></td><td> : Go to export</td></tr>
	      <tr><td align="right"><b>g</b> then <b>g</b></td><td> : Go to top of page</td></tr>
	      <tr><td align="right"><b>g</b> then <b>b</b></td><td> : Go to bottom of page</td></tr>
	      <tr><td align="right"><b>g</b> then <b>i</b></td><td> : Go to index</td></tr>
	      <tr><td align="right"><b>g</b> then <b>t</b></td><td> : Add or remove labels</td></tr>
	      <tr><td align="right"><b>g</b> then <b>e</b></td><td> : Add examples</td></tr>
          </tbody>
      </table>
  </div>

  <div class="modal-footer">
      <a href="#" class="btn" data-dismiss='modal' aria-hidden="true">Close</a>
      <!--<a href="#" class="btn btn-primary">Save changes</a>-->
  </div>
</div>

<footer id="_bottom">
	<div class="container" style="padding-top: 10px; padding-bottom: 10px; width: 1050px;">
		<div id="footer" class="span6" style="width: {{if .IsHome}}440{{else}}600{{end}}px;">
			<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.10.1/jquery.min.js"></script>
			<script type="text/javascript" src="/static/js/bootstrap.min.js"></script>
			<script type='text/javascript' src='/static/js/site.js'></script>

			<p><strong>Copyright © 2013 Go Walker</strong></p>
			<p>Website built by <a target="_blank" href="https://github.com/Unknwon"><i class="icon-user"></i> @Unknown</a>. Powered by <a target="_blank" href="https://github.com/astaxie/beego"><i class="icon-bold"></i>eego</a> and <a target="_blank" href="https://github.com/coocood/qbs">Qbs</a>.</p>
			{{if .IsHome}}<p>Based on <a target="_blank" href="http://twitter.github.io/bootstrap/">Bootstrap</a>. Icons from <a target="_blank" href="http://glyphicons.com/">Glyphicons</a>.</p>{{end}}
			{{if .IsHome}}<p>{{end}}Send us <a href="mailto:joe2010xtmf#163.com"><i class="icon-envelope"></i> Feedback</a> or submit <a target="_blank" href="https://github.com/Unknwon/gowalker/issues"><i class="icon-tasks"></i> Website Issues</a>. {{if .IsHome}}</p>{{end}}<strong>Language:</strong>

		    <div class="btn-group dropup">
			    <button class="btn dropdown-toggle" data-toggle="dropdown">{{.CurLang}} <span class="caret"></span></button>
			    <ul class="dropdown-menu">
				{{$keyword := .Keyword}}
			    	{{range .RestLangs}}
			    	<li><a href="?lang={{.Lang}}&q={{$keyword}}">{{.Name}}</a></li>
			    	{{end}}
			    </ul>
		    </div>
		   {{template "static_file" .}}
	</div>
</footer>
{{end}}

{{define "footer_zh"}}
<div id='_keyshortcut' class="modal hide fade" tabindex="-1" role="dialog" aria-hidden="false">
  <div class="modal-header">
      <h3>控制面板</h3>
  </div>

  <div class="modal-body">
      <table>
          <tbody>
	      <tr><td align="right"><b>?</b></td><td> ： 打开控制面板</td></tr>
	      <tr><td align="right"><b>/</b></td><td> ： 站内搜索</td></tr>
	      <tr><td align="right"><b>.</b></td><td> ： 搜索导出对象</td></tr>
	      <tr><td align="right"><b>g</b> + <b>g</b></td><td> ： 回到顶部</td></tr>
	      <tr><td align="right"><b>g</b> + <b>b</b></td><td> ： 跳至底部</td></tr>
	      <tr><td align="right"><b>g</b> + <b>i</b></td><td> ： 跳至索引</td></tr>
	      <tr><td align="right"><b>g</b> + <b>t</b></td><td> ： 增加或移除标签</td></tr>
	      <tr><td align="right"><b>g</b> + <b>e</b></td><td> ： 增加示例</td></tr>
          </tbody>
      </table>
  </div>

  <div class="modal-footer">
      <a href="#" class="btn" data-dismiss='modal' aria-hidden="true">关闭</a>
      <!--<a href="#" class="btn btn-primary">Save changes</a>-->
  </div>
</div>

<footer id="_bottom">
	<div class="container" style="padding-top: 10px; padding-bottom: 10px; width: 1050px;">
		<div id="footer" class="span6" style="width: {{if .IsHome}}440{{else}}600{{end}}px;">
			<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.10.1/jquery.min.js"></script>
			<script type="text/javascript" src="/static/js/bootstrap.min.js"></script>
			<script type='text/javascript' src='/static/js/site.js'></script>

			<p><strong>版权所有 © 2013 Go 步行者</strong></p>
			<p>网站建设者 <a target="_blank" href="https://github.com/Unknwon"><i class="icon-user"></i> @Unknown</a>. 项目基于 <a target="_blank" href="https://github.com/astaxie/beego"><i class="icon-bold"></i>eego</a> 和 <a target="_blank" href="https://github.com/coocood/qbs">Qbs</a> 构建。</p>
			{{if .IsHome}}<p>界面基于 <a target="_blank" href="http://twitter.github.io/bootstrap/">Bootstrap</a>. 图标来自 <a target="_blank" href="http://glyphicons.com/">Glyphicons</a>.</p>{{end}}
			{{if .IsHome}}<p>{{end}}发送 <a href="mailto:joe2010xtmf#163.com"><i class="icon-envelope"></i> 反馈</a> 或者提交 <a target="_blank" href="https://github.com/Unknwon/gowalker/issues"><i class="icon-tasks"></i> 网站问题</a>. {{if .IsHome}}</p>{{end}}<strong>语言选项：</strong>
		    <div class="btn-group dropup">
			    <button class="btn dropdown-toggle" data-toggle="dropdown">{{.CurLang}} <span class="caret"></span></button>
			    <ul class="dropdown-menu">
				{{$keyword := .Keyword}}
			    {{range .RestLangs}}
			    	<li><a href="?lang={{.Lang}}&q={{$keyword}}">{{.Name}}</a></li>
			    {{end}}
			    </ul>
		    </div>
		   {{template "static_file" .}}
	</div>
</footer>
{{end}}

{{define "static_file"}}
	<!-- <span class="muted">|</span>
	<span>
		<script type="text/javascript">
			var _bdhmProtocol = (("https:" == document.location.protocol) ? " https://" : " http://");
			document.write(unescape("%3Cscript src='" + _bdhmProtocol +"hm.baidu.com/h.js%3Fd2d5278d61e466bcb3f9ea29a18d40dc' type='text/javascript'%3E%3C/script%3E"));
		</script>
	</span>
	<span class="muted">|</span>
	<span>
		<script type="text/javascript" src="http://tajs.qq.com/stats?sId=24262957" charset="UTF-8"></script>
	</span> -->
</div>

{{if .IsHome}}
<div class="span6" style="margin-top: 25px;">
	<div style="text-align: center;">
		<img src="/static/img/qiniu.png">
	</div>

	<div style="text-align: center;margin-top: 15px;">
		<img src="/static/img/Golang.png">
		<img src="/static/img/bee.gif" style="width: 58px;">
		<img src="/static/img/linode.png">
	</div>
</div>

<script >
	// Call popover for features.
	$(function () {
	    $('.feature').popover()
	})
</script>
{{else}}
{{if .IsHasReadme}}
<script type="text/javascript" src="/static/js/showdown.js"></script>
<script type="text/javascript">
	var converter = new Showdown.converter({});
	var readme = document.getElementById("readme");
	var content = converter.makeHtml(document.getElementById("readme").innerHTML);
	content = content.replace(/&amp;/g, "&");
	readme.innerHTML = content;
	readme.style.display = "block";
</script>
{{end}}

<script type="text/javascript">
    (function($){

        var doc = $("#markdown");

        // Set anchor.
        doc.find("h1, h2, h3, h4, h5, h6").each(function(){
            var node = $(this);
            var id = node.attr("id");

            if (id != "undefined") {
	            node = node.wrap('<div class="anchor-wrap" ></div>');
	            node.append('<a class="anchor" href="#' + node.attr("id") + '"><span class="octicon octicon-link"></span></a>')
        	}
        });

    })(jQuery);
</script>
{{end}}

<script>
	// (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
	// (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
	// m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
	// })(window,document,'script','//www.google-analytics.com/analytics.js','ga');

	// ga('create', 'UA-40109089-2', 'gowalker.org');
	// ga('send', 'pageview');
</script>
{{end}}