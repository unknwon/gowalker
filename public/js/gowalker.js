$(document).ready(function () {
    $('.dropdown').dropdown({
        transition: 'drop'
    });
    $('.popup').popup();
    $('.ui.accordion').accordion();
    $('.ui.search').search({
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

    $('.ex-link').click(function () {
        $($(this).data("name")).addClass("active");
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