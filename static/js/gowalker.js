$(document).ready(function () {
    $('.ui.dropdown').dropdown();
    $('.ui.feature').popup();

    var $backToTopTxt = "Back to Top",
        $backToTopEle = $('<div class="backToTop"></div>').appendTo($("body")).attr("title", $backToTopTxt).click(function () {
            $("html, body").animate({ scrollTop: 0 }, 120);
        }), $backToTopFun = function () {
            var st = $(document).scrollTop(), winh = $(window).height();
            (st > 0) ? $backToTopEle.show() : $backToTopEle.hide();
            //IE6下的定位
            if (!window.XMLHttpRequest) {
                $backToTopEle.css("top", st + winh - 166);
            }
        };
    $(window).bind("scroll", $backToTopFun);
    $backToTopFun();

    // Home page.
    if ($('#home-search-form').length) {
        $('#search-btn').click(function () {
            $('#home-search-form').submit();
        });

        var now = new Date();
        $("table[name='history-table']").each(function () {
            $(this).find('.meta-time').each(function () {
                $(this).text(new Date(parseInt($(this).text()) * 1000).toLocaleString());
            });
        });
    }
});