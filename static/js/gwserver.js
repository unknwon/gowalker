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
        var navbarFrame = document.getElementById("navbar_frame");
        var body = document.getElementById("body");
        var delta = document.body.clientWidth - 1050;

        if (delta >= 0) {
            if (document.getElementById("sidebar") == null) {
                navbarFrame.className = "navbar navbar-default navbar-fixed-top";
                body.style.paddingTop = "60px";
            } else {
                navbarFrame.className = "navbar navbar-default";
                body.style.paddingTop = "0px";
            }

        } else {
            navbarFrame.className = "navbar navbar-default";
            body.style.paddingTop = "0px";
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

    if (document.body.clientWidth > 1500 && document.getElementById("sidebar") != null) {
        document.getElementById("sidebar").className = "span3"
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

    // For label modal.
    var _tf = $('#label_modal');
    if (_tf.length != 0) {
        _tf.modal({ keyboard: false, show: false }); // For tags modal.

        $('#label_form').submit(function () {
            var _tf = $('#label_modal');
            _tf.modal('hide');
            _tf.find('input[type=text]').val("");
            return false;
        });
    } else {
        _tf = null;
    }

    // For example modal.
    var _ex = $('#example_modal');
    if (_ex.length != 0) {

    } else {
        _ex = null;
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
            if (_ex) _ex.modal('hide');
            _modal.modal('show');
        } else if (code == 47) { //for '/'    forward slash code:47
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            if (_ex) _ex.modal('hide');
            _modal.modal('hide');
            //site search focus
            $('input[name=q]').first().focus();
            return false;
        } else if (code == 46 && isProjectPage) { //for '.'    comma as 46   'go to export'
            if (_tf) _tf.modal('hide');
            if (_ex) _ex.modal('hide');
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
            if (_ex) _ex.modal('hide');
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
            if (_ex) _ex.modal('hide');
            _modal.modal('hide');
            GkeyCb(function () {
                $("html,body").animate({ scrollTop: $("body").height() }, 120);
            });

        } else if (code == 105) {//for 'g then i'     i  105
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            if (_ex) _ex.modal('hide');
            _modal.modal('hide');
            GkeyCb(function () {
                location.hash = "#_index";
            });
        } else if (code == 116) { // for `g then t`	t 	116
            if (_ep) _ep.modal('hide');
            if (_ex) _ex.modal('hide');
            _modal.modal('hide');
            GkeyCb( function(){
                    if (_tf) {
                        _tf.modal('show');
                        _tf.on('shown', function () {
                        $(this).find('#label_box').focus();
                    })
                }
            });
        } else if (code == 101) {// for 'g then e'   e 101
            if (_ep) _ep.modal('hide');
            if (_tf) _tf.modal('hide');
            _modal.modal('hide');
            GkeyCb( function() {
                if (_ex) {
                    _ex.modal('show');
                    _ex.on('shown', function () {
                        $(this).find('#example_box').focus();
                    })
                }
            });
        }
    })
    //end
})();

// Add tag.
function AddLabelSubmit(obj) {
    var input = $.trim(document.getElementById("label_box").value);
    if (input.length > 0) {
        var anchor;
        if (window.location.href.indexOf("?") > -1) {
            anchor = window.location.href.replace("?", ":l=" + input + "?");
        } else {
            anchor = window.location.href + ":l=" + input;
        }
        window.location.href = anchor;
        return false;
    }
}

// Remove tag.
function RemoveLabelSubmit(obj) {
    var input = $.trim(document.getElementById("label_box").value);
    if (input.length > 0) {
        var anchor;
        if (window.location.href.indexOf("?") > -1) {
            anchor = window.location.href.replace("?", ":rl=" + input + "?");
        } else {
            anchor = window.location.href + ":rl=" + input;
        }
        window.location.href = anchor;
        return false;
    }
}

function showExample(name) {
    var ex = document.getElementById("_ex_" + name);
    ex.className = "accordion-body collapse in";
    ex.style.height = "auto";
}

function TransToChinese(node) {
    switch (node.attr("id")) {
        case "_index":
            node.html(node.html().toString().replace("Index", "索引"));
            break;
        case "_exams":
            node.html("示例");
            break;
        default:
            node.html(node.html().toString().replace("View Code", "查看代码"));
    }
}

function viewCode(name) {
    var decl = document.getElementById(name.replace("|||", "_"));
    if (decl.style.borderBottom == "0") {
        decl.style.borderBottom = "1px";
    } else {
        decl.style.borderBottom = "0px";
    }
}
