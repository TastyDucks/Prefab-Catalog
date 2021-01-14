$(document).ready(function() {
    $('#dataTable').DataTable( {
        columnDefs: [
            {
                targets: '_all',
                className: 'dt-body-center'
            }
        ],
        paging:   false,
        fixedHeader: true,
    } );
} );
