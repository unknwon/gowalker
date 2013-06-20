/**
 *  @create 2013/06/11
 *  @version 0.2
 *  @author: chenwenli <kapa2robert@gmail.com>, Unknown <joe2010xtmf@163.com>
 */

(function () {
    // Fit navbar padding.
    Responsive();
    $(window).resize(function () {
        Responsive();
    });

    function Responsive() {
        var navbar = document.getElementById("navbar");
        var searchForm = document.getElementById("top_search_form");
        var searchBox = document.getElementById("navbar_search_box");
        var homeBody = document.getElementById("home_content");
        var footer = document.getElementById("footer");
        var delta = document.body.clientWidth - 1100;

        if (delta > 0) {
            navbar.style.paddingLeft = delta / 2 + 40 + "px";
            navbar.style.paddingRight = delta / 2 + 70 + "px";
            searchBox.style.width = "";

            // Home page.
            if (homeBody != null) {
                homeBody.style.marginLeft = "-50px";
            }

            footer.style.marginLeft = "0px";
        } else {
            navbar.style.paddingLeft = "30px";
            navbar.style.paddingRight = "10px";

            if (document.body.clientWidth < 700) {
                searchForm.className = "navbar-search pull-right hide";
            } else {
                searchForm.className = "navbar-search pull-right";
                searchBox.style.width = "150px";
            }

            // Home page.
            if (homeBody != null) {
                homeBody.style.marginLeft = "0px";
            }

            footer.style.marginLeft = "20px";
        }
    }

    var $backToTopTxt = "Back to Top", $backToTopEle = $('<div class="backToTop"></div>').appendTo($("body")).attr("title", $backToTopTxt).click(function () {
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

    if (document.body.clientWidth > 1200 && document.getElementById("sidebar") != null) {
        document.getElementById("sidebar").className = ""
    }

    var _ep = $('#search_exports');
    if (_ep.length != 0) {
        _ep.modal({ keyboard: false, show: false }); // for export modal

        $('#search_form').submit(function () {
            var input = $.trim(document.getElementById("search_export_box").value)
            if (input.length > 0) {
                _ep.modal('hide');
                var anchor = "#".concat(input.replace(".", "_"));
                location.hash = anchor;
            }
            _ep.find('input[type=text]').val("");
            return false;
        });
    } else {
        _ep = null;
    }

    // For tags modal.
    var _tf = $('#tags_modal');
    if (_tf.length != 0) {
        _tf.modal({ keyboard: false, show: false }); // For tags modal.

        $('#tags_form').submit(function () {
            var _tf = $('#tags_modal');
            _tf.modal('hide');
            _tf.find('input[type=text]').val("");
            return false;
        });
    } else {
        _tf = null;
    }

    // For global modal.
    var _modal = $("#_keyshortcut");
    _modal.modal({ keyboard: true, show: false });
    var isProjectPage = 0;
    var preKeyG = 0;
    if (document.getElementById("sidebar") != null) {
        isProjectPage = 1;

    } else {
        // Mute options in control panel.
        _modal.find('tbody > tr').each(function (i, ele) {
            if (i == 2 || i == 5 || i == 6 || i == 7) {
                $(ele).addClass("muted");
            }
        })
    }

    function GkeyCb(callback) {
        if (preKeyG == 1) {
            callback();
        }
        preKeyG = 0;
    }

    $(document).keypress(function (event) {
        if ($('input:focus').length != 0) {
            return true;
        }
        var code = event.keyCode ? event.keyCode : event.charCode;
        if (code == 63) {// for '?'  equal as  63
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            _modal.modal('show');
        } else if (code == 47) { //for '/'    forward slash code:47
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            _modal.modal('hide');
            //site search focus
            $('input[name=q]').first().focus();
            return false;
        } else if (code == 46 && isProjectPage) { //for '.'    comma as 46   'go to export'
            if (_tf) _tf.modal('hide');
            _modal.modal('hide');
            if (_ep) {
                _ep.modal('show');
                _ep.on('shown', function () {
                    $(this).find('#search_export_box').focus();
                })
            }
        } else if (code == 103) {// for 'g then g'   g 103
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            _modal.modal('hide');
            if (preKeyG == 0) {
                preKeyG = 1;
                setTimeout(function () {
                    preKeyG = 0
                }, 2000);
                return false;
            }
            //                           console.log(preKeyG);
            GkeyCb(function () {
                $("html,body").animate({ scrollTop: 0 }, 120);
            });

        } else if (code == 98) {//for 'g then b'    b 98
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            _modal.modal('hide');
            GkeyCb(function () {
                $("html,body").animate({ scrollTop: $("body").height() }, 120);
            });

        } else if (code == 105) {//for 'g then i'     i  105
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            _modal.modal('hide');
            GkeyCb(function () {
                location.hash = "#_index";
            });
        } else if (code == 116) { // for `g then t`	t 	116
            if (_ep) _ep.modal('hide');
            _modal.modal('hide');
            if (_tf) {
                _tf.modal('show');
                _tf.on('shown', function () {
                    $(this).find('#tags_box').focus();
                })
            }
        } else if (code == 101) {// for 'g then e'   e 101
            if (_ep) _ep.modal('hide');
            _modal.modal('hide');
            GkeyCb(function () {
                location.hash = "#Chunk";
            });
        }
    })
    //end
})();

// Add tag.
function AddTagSubmit(obj) {
    var input = $.trim(document.getElementById("tags_box").value);
    if (input.length > 0) {
        var anchor;
        if (window.location.href.indexOf("?") > -1) {
            anchor = window.location.href.replace("?", ":tag=" + input + "?");
        } else {
            anchor = window.location.href + ":tag=" + input;
        }
        window.location.href = anchor;
        return false;
    }
}

// Remove tag.
function RemoveTagSubmit(obj) {
    var input = $.trim(document.getElementById("tags_box").value);
    if (input.length > 0) {
        var anchor;
        if (window.location.href.indexOf("?") > -1) {
            anchor = window.location.href.replace("?", ":rtag=" + input + "?");
        } else {
            anchor = window.location.href + ":rtag=" + input;
        }
        window.location.href = anchor;
        return false;
    }
}
//end