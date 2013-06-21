{{define "footer_en"}}
<!DOCTYPE html>
		<div id='_keyshortcut' class="modal hide fade" tabindex="-1" role="dialog" aria-hidden="false">
		  <div class="modal-header">
		      <h3>Keyboard Shortcuts</h3>
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
			      <tr><td align="right"><b>g</b> then <b>l</b></td><td> : Add or remove labels</td></tr>
			      <tr><td align="right"><b>g</b> then <b>e</b></td><td> : Go to examples</td></tr>
		          </tbody>
		      </table>
		  </div>
		  <div class="modal-footer">
		      <a href="#" class="btn" data-dismiss='modal' aria-hidden="true">Close</a>
		      <!--<a href="#" class="btn btn-primary">Save changes</a>-->
		  </div>
		</div>
		<footer>
			<div class="container" style="padding-top: 10px; padding-bottom: 10px; width: 1050px;">
				<div id="footer" class="span6" style="width: 440px;">
					<script type="text/javascript" src="/static/js/jquery.js"></script>
					<script type="text/javascript" src="/static/js/bootstrap.min.js"></script>
					<script type='text/javascript' src='/static/js/site.js'></script>
					<p><strong>Copyright © 2013 Go Walker</strong></p>
					<p>Website built by <a target="_blank" href="https://github.com/Unknwon"><i class="icon-user"></i> @Unknown</a>. Powered by <a target="_blank" href="https://github.com/astaxie/beego"><i class="icon-bold"></i>eego</a>, <a target="_blank" href="https://github.com/coocood/qbs">Qbs</a>, <a target="_blank" href="https://github.com/mattn/go-sqlite3">go-sqlite3</a>.</p>
					<p>Based on <a target="_blank" href="http://twitter.github.io/bootstrap/">Bootstrap</a>. Icons from <a target="_blank" href="http://glyphicons.com/">Glyphicons</a>.</p>
					<p>Send us <a href="mailto:joe2010xtmf#163.com"><i class="icon-envelope"></i> Feedback</a> or submit <a target="_blank" href="https://github.com/Unknwon/gowalker/issues"><i class="icon-tasks"></i> Website Issues</a>.</p><strong>Language:</strong>
				    <div class="btn-group dropup">
					    <button class="btn dropdown-toggle" data-toggle="dropdown">{{.CurLang}} <span class="caret"></span></button>
					    <ul class="dropdown-menu">
						{{$keyword := .Keyword}}
					    {{range .RestLangs}}
					    	<li><a href="?lang={{.Lang}}&q={{$keyword}}">{{.Name}}</a></li>
					    {{end}}
					    </ul>
				    </div>
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
				{{end}}
				<script>
				  // (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
				  // (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
				  // m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
				  // })(window,document,'script','//www.google-analytics.com/analytics.js','ga');

				  // ga('create', 'UA-40109089-2', 'gowalker.org');
				  // ga('send', 'pageview');
				</script>
			</div>
		</footer>
{{end}}