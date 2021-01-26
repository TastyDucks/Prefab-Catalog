$(document).ready(function() {
    var table = $('#dataTable').DataTable( {
        columnDefs: [
            {
                targets: '_all',
                className: 'dt-body-center'
            }
        ],
        paging:   false,
        fixedHeader: true,
    } );
    // Simulate clearing any search so that we can get the POST quantity fields that DataTables may have removed from the DOM.
    $('button').click( function() {
        var data = table.$('input').serialize();
        const e = $.Event('paste');
        $('[aria-controls="dataTable"]').val('').trigger(e);
        return true;
    } );
} );
