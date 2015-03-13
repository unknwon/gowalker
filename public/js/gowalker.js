$(document).ready(function () {
    $('.dropdown').dropdown({
        transition: 'drop'
    });
    $('.popup').popup();
    $('.ui.accordion').accordion();
    $('.ui.main.search').search({
        type: "standard",
        apiSettings: {
            url: '/search/json?q={query}'
        },
        searchFields: [
            'title'
        ],
        searchDelay: 500,
        searchFullText: false
    });
    $('.main.search .search.icon').click(function () {
        $('.main.search').submit();
    });

    // View code.
    $('.show.code').click(function () {
        $($(this).data('target')).toggle();
    });

    // Show example.
    $('.ex-link').click(function () {
        $($(this).data("name")).show();
    });
    $('.show.example').click(function (event) {
        $($(this).attr('href')).toggle();
        event.preventDefault();
    });


    // Browse history.
    if ($('#browse_history').length) {
        $(this).each(function () {
            $(this).find('.meta-time').each(function () {
                $(this).text(new Date(parseInt($(this).text()) * 1000).toLocaleString());
            });
        });
    }

    // Resize images if too large.
    if ($('#readme').length) {
        $(this).find("img").load(function () {
            var w = $(this).width();
            $(this).width(w > 600 ? 600 : w);
        });
    }

    // Anchor.
    if ($('#markdown').length) {
        $(this).find('h1, h2, h3, h4, h5, h6').each(function () {
            var node = $(this);
            var id = node.attr("id");
            if (typeof(id) !== "undefined") {
                node = node.wrap('<div class="anchor-wrap" ></div>');
                node.append('<a class="anchor" href="#' + node.attr("id") + '"><span class="octicon octicon-link"></span></a>');
            }
        });
    }
});