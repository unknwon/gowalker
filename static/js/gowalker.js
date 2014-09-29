$(document).ready(function () {
    $('.ui.dropdown').dropdown();
    $('.ui.feature').popup();

    $('#search-btn').click(function () {
        $('#home-search-form').submit();
    });

    // Home page.
    if ($('#home-search-form').length) {
        var now = new Date();
        $("table[name='history-table']").each(function () {
            $(this).find('.meta-time').each(function(){
                $(this).text(new Date(parseInt($(this).text()) * 1000).toLocaleString());
            });
        });
    }
});