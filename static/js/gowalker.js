$(document).ready(function () {
	$('.ui.dropdown').dropdown();
	$('.ui.feature').popup();
	
	$('#search-btn').click(function(){
		$('#main-search-form').submit();
	});
});