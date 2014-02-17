<div id="example_modal" class="modal fade" tabindex="-1">
    <div class="modal-dialog">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button>
                <h3 style="margin: 10px">{{i18n .Lang "Add examples"}}.</h3>
            </div>
            <form id="example_form" class="modal-form" action="/example">
                <input type="hidden" name="q" value="{{.ImportPath}}">	
                <div class="modal-body">
                    <input id="example_box" autofocus="autofocus" type="text" name="gist" placeholder="{{i18n .Lang "Please type Gist address"}}">
                </div>
                
                <div class="modal-footer">
                    <button type='button' class="btn-default" data-dismiss="modal" aria-hidden="true">{{i18n .Lang "Close"}}</button>
                    <button class="btn btn-primary" type="submit">{{i18n .Lang "Gist!"}}</button>
                </div>
            </form>
        </div>
    </div>
</div>