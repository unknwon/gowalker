$(document).ready(function () {
    $('.dropdown').dropdown({
        transition: 'drop'
    });
    $('.popup').popup();
    $('.ui.accordion').accordion();

    function setSearchOptions(semantic_search) {
        $('.ui.main.search').search({
            type: "standard",
            minCharacters: 3,
            apiSettings: {
                url: '/search/json?q={query}&semantic_search=' + semantic_search,
                onResponse: function (json) {
                    var response = {
                        results: []
                    };

                    var maxResults = 7;
                    $.each(json.results, function (index, item) {
                        if (index > maxResults) {
                            return false;
                        }

                        response.results.push({
                            title: item.title,
                            description: item.description,
                            url: item.url
                        });
                    });

                    return response;
                }
            },
            searchDelay: 500,
            searchFullText: false
        });
    }

    setSearchOptions(localStorage.enable_semantic_search);
    if (localStorage.enable_semantic_search == 'true') {
        $('#semantic_search_checkbox').checkbox('check');
    }

    $('.main.search .search.icon').click(function () {
        $('.main.search').submit();
    });

    // Toggle semantic search
    $('#semantic_search').change(function () {
        if (this.checked) {
            setSearchOptions(true);
        } else {
            setSearchOptions(false);
        }
        $('.ui.main.search').search('clear cache', $('#search_input').value)
        localStorage.enable_semantic_search = this.checked;
    });

    var is_page_docs = $('#readme').length > 0;

    if (is_page_docs) {
        // Search export objects.
        var $searchExportPanel = $('.search.export.panel');
        var $searchExportForm = $('.ui.form.search.export');
        var $searchExportInput = $('.ui.form.search.export input');
        $searchExportForm.search({source: exportDataSrc});
        $searchExportForm.submit(function (event) {
            $searchExportPanel.modal("hide");
            window.location.href = "#" + $searchExportInput.val().replace(/\./g, "_");
            event.preventDefault();
        });
    }

    // Control panel.
    var $control_panel = $('.control.panel');
    $('#control-panel').click(function (event) {
        $(this).blur();
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
        // Check if any input box is focused.
        if ($(':focus').length > 0) {
            return true;
        }

        var code = event.keyCode ? event.keyCode : event.charCode;
        switch (code) {
            case 63:                    // for '?' 63
                $control_panel.modal('show');
                break;
            case 98:                    // for 'g then b'  'b' 98
                $control_panel.modal('hide');
                Gkey(function () {
                    $('html,body').animate({scrollTop: $(document).height()}, 120);
                });
                break;
            case 103:                   // for 'g then g'   'g' 103
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
                break;
            case 115:                   // for 's' 105
                if (!is_page_docs)return true;
                $searchExportPanel.modal('show');
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