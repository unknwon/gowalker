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

    // Control panel.
    var $control_panel = $('.control.panel');
    $('#control-panel').click(function (event) {
        $control_panel.modal('show');
        event.preventDefault();
    });

    var preKeyG = 0;

    function Gkey(callback) {
        if (preKeyG === 1) {
            callback();
        }
        preKeyG = 0;
    }

    $(document).keypress(function (event) {
        var code = event.keyCode ? event.keyCode : event.charCode;
        if (code === 63) {              // for '?' 63
            $control_panel.modal('show');
        } else if (code === 103) {      // for 'g then g'   'g' 103
            $control_panel.modal('hide');
            if (preKeyG === 0) {
                preKeyG = 1;
                setTimeout(function () {
                    preKeyG = 0;
                }, 2000);
                return false;
            }
            Gkey(function () {
                $('html,body').animate({scrollTop: 0}, 120);
            });
        } else if (code === 98) {        // for 'g then b'  'b' 98
            $control_panel.modal('hide');
            Gkey(function () {
                $('html,body').animate({scrollTop: $(document).height()}, 120);
            });
        } else {
            preKeyG = 0;
        }
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