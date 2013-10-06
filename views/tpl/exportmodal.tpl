<div id="search_exports" class="modal fade" tabindex="-1">
    <div class="modal-dialog">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>
                <h3 id="search_export_title" style="margin: 10px">Search and Go to Exports.</h3>
            </div>
            
            <form id="search_form" class="modal-form">
                <div class="modal-body">
                    <input id="search_export_box" autofocus="autofocus" autocomplete="off" class="search-export" style="width: 400px;" type="text" placeholder="Input type, method or function name">
                </div>
                    
                {{str2html .ExportDataSrc}}
                
                <div class="modal-footer">
                    <button id="close_button" type='button' class="btn" data-dismiss="modal" aria-hidden="true">Close</button>
                    <button id="search_export_button" class="btn btn-primary" type="submit">Go!</button>
                </div>
                    
                <script>
                    $('.tt-query').css('background-color','#fff');
                </script>
            </form>
        </div>
    </div>
</div>