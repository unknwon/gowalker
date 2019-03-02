$(document).ready(function () {
    $('.ui.accordion .title').click(function () {
        var $icon = $(this).find('.fas');
        var $content = $('.ui.accordion .content');
        if ($content.hasClass('d-hide')) {
            // Resize images if too large.
            if ($('#readme').length) {
                $(this).find("img").each(function () {
                    var w = $(this).width();
                    $(this).width(w > 600 ? 600 : w);
                });
            }

            $content.removeClass('d-hide');
            $icon.removeClass('fa-caret-right');
            $icon.addClass('fa-caret-down');
            return;
        }

        $content.addClass('d-hide');
        $icon.removeClass('fa-caret-down');
        $icon.addClass('fa-caret-right');
    });

    var delay = (function(){
        var timer = 0;
        return function(callback, ms){
          clearTimeout(timer);
          timer = setTimeout(callback, ms);
        };
    })();

    $('#import-search').keyup(function() {
        delay(function () {
            var $this = $('#import-search');
            if ($this.val().length < 3) {
                $('#search-results').html("");
                return;
            }

            $('.form-autocomplete i.fa-search').addClass('d-hide');
            $('.form-autocomplete .loading').removeClass('d-hide');
            $.get('/search/json?q=' + $this.val(), function (data) {
                $('#search-results').html("");
                for (var i = 0; i < data.results.length; i++) {
                    $('#search-results').append(`<a href="` + data.results[i].url + `">
                    <div class="tile tile-centered">
                        <div class="tile-content">
                            <p class="tile-title"><b>` + data.results[i].title + `</b></p>
                            <p class="tile-subtitle">` + data.results[i].description + `</p>
                        </div>
                    </div>
                </a>`)
                    if (i+1 < data.results.length) {
                        $('#search-results').append(`<div class="divider"></div>`)
                    }
                }
                $('.form-autocomplete .loading').addClass('d-hide');
                $('.form-autocomplete i.fa-search').removeClass('d-hide');
            });
        }, 500)
    });

    var is_page_docs = $('#readme').length > 0;

    if (is_page_docs) {
        // Search export objects.
        var $searchExportPanel = $('#search-export-panel');
        $searchExportPanel.find('.btn-clear').click(function () {
            $searchExportPanel.removeClass('active');
        })
        $('#exports-search').keyup(function () {
            if ($(this).val().length < 1) {
                $('#search-results').html("");
                return;
            }

            $('#search-results').html("");
            for (var i = 0; i < exportDataSrc.length; i++) {
                if (exportDataSrc[i].title.toLowerCase().includes($(this).val().toLowerCase())) {
                    $('#search-results').append(`<a href="#` + exportDataSrc[i].title.replace(/\./g, "_") + `">
                    <div class="tile tile-centered">
                        <div class="tile-content">` + exportDataSrc[i].title + `</div>
                    </div>
                </a>`)
                }
            }

            $('#search-results a').click(function () {
                $searchExportPanel.removeClass('active');
            });
        });
    }

    // Help panel
    var $help_panel = $('#help-panel');
    $help_panel.find('.btn-clear').click(function () {
        $help_panel.removeClass('active');
    })

    var preKeyG = 0;

    function Gkey(callback) {
        if (preKeyG === 1) {
            callback();
        }
        preKeyG = 0;
    }

    $(document).keypress(function (event) {
        // Check if any input box is focused.
        if ($('input:focus').length > 0) {
            return true;
        }

        var code = event.keyCode ? event.keyCode : event.charCode;
        switch (code) {
            case 63:                    // for '?' 63
                $help_panel.addClass('active');
                break;
            case 98:                    // for 'g then b'  'b' 98
                $help_panel.removeClass('active');
                Gkey(function () {
                    $('html,body').animate({scrollTop: $(document).height()}, 120);
                });
                break;
            case 103:                   // for 'g then g'   'g' 103
                $help_panel.removeClass('active');

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
                break;
            case 115:                   // for 's' 105
                if (!is_page_docs) return true;
                $searchExportPanel.addClass('active');
                break;
            default:
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