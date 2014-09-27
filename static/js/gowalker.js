$(document).ready(function () {
	$('.ui.dropdown').dropdown();
	$('#search-btn').click(function(){
		$('#main-search-form').submit();
	});
});