/**
 *	@create 2013/06/11
 *  @version 0.1 
 */

//add by chenwenli <kapa2robert@gmail.com>
(function() {     
	var $backToTopTxt = "Back to Top",$backToTopEle = $('<div class="backToTop"></div>').appendTo($("body")).attr("title", $backToTopTxt).click(function() {             
		$("html, body").animate({ scrollTop: 0 }, 120);     
	}), $backToTopFun = function() {         
		var st = $(document).scrollTop(), winh = $(window).height();         
		(st > 0)? $backToTopEle.show(): $backToTopEle.hide();        
		//IE6下的定位         
		if (!window.XMLHttpRequest) {             
			$backToTopEle.css("top", st + winh - 166);         
		}     
	};     
	$(window).bind("scroll", $backToTopFun);

	$backToTopFun(); 

	if (document.body.clientWidth > 1500 && document.getElementById("navbar") != null )
	{
		document.getElementById("navbar").className=""
	}  

	var _ep = $('#search_exports');
	if(_ep.length != 0){
		_ep.modal({ keyboard: false, show: false }); // for export modal

		$('#search_form').submit(function(){
			var input = $.trim(document.getElementById("search_export_box").value)
			if(input.length > 0 ){
				_ep.modal('hide');
	 			var anchor = "#".concat(input.replace(".", "_"));
	 			if(location.hash == anchor){
	 				location.hash = "";
	 			}
	 			location.hash = anchor;	
			}
			return false;
		});
	}else{
		_ep = null;
	}
	

	//for global modal 
   	var _modal = $("#_keyshortcut");
    _modal.modal({ keyboard: true, show: false });
   	  
    var isProjectPage = 0;
    var preKeyG = 0;
    	if(  document.getElementById("navbar") != null){
    		isProjectPage = 1;

    	}else{
    		// Mute options in control panel.
    		_modal.find('tbody > tr').each(function(i,ele){
    			if(i == 2 || i ==5 || i ==6){
    				$(ele).addClass("muted");
    			}
    		})
    	}

    function  GkeyCb(callback){
    	if(preKeyG ==1 ){
       	  	callback();
       	}
       	preKeyG = 0;
    }

	$(document).keypress(function(event){
		if($('input:focus').length != 0){
			return true;
		}
		var code = event.keyCode ? event.keyCode : event.charCode;
		if(code == 63 ){// for '?'  equal as  63
		    if(_ep) _ep.modal('hide');
		    _modal.modal('show');
		}else if(code == 47){ //for '/'    forward slash code:47
				if(_ep) _ep.modal('hide');
		    _modal.modal('hide');
		    //site search focus
		    $('input[name=q]').first().focus();
		    $('input[name=q]').first().html("");
		}else if( code == 46 && isProjectPage){ //for '.'    comma as 46   'go to export'
				_modal.modal('hide');  
				if(_ep) _ep.modal('show');
		}else if( code == 103){// for 'g then g'   g 103
				if(_ep) _ep.modal('hide');
				_modal.modal('hide');
				if(preKeyG == 0 ){
					preKeyG  =1;
					setTimeout(function(){ preKeyG = 0 }, 2000);
					return false;
		   		}
		//                           console.log(preKeyG);
				GkeyCb(function(){
					$("html,body").animate({ scrollTop: 0 }, 120);
				});
				
		}else if( code ==  98){//for 'g then b'    b 98
				if(_ep) _ep.modal('hide');
				_modal.modal('hide');
				GkeyCb(function(){
					$("html,body").animate({ scrollTop :$("body").height() } ,120);
		   		});

	    }else if( code ==  105){//for 'g then i'     i  105
	     		if(_ep) _ep.modal('hide');
	     		_modal.modal('hide');
	     		GkeyCb(function(){
	     			location.hash = "#_index" ;
	     		});
	    }
	    // else if( code == 101 ){// for 'g then e'   e 101
	    //  		if(_ep) _ep.modal('hide');
	    //  		_modal.modal('hide');
	    //  		GkeyCb(function(){
	    //  			location.hash = "#Chunk";
	    //  		});
	    // }
		// else if(code == ){}    // for `g then t`   
	})
	//end 
})();
//end