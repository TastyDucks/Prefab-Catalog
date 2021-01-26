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
        initComplete: function () {
            this.api().columns([2, 3, 4, 5]).every( function () {
                var column = this;
                var select = $('<select><option value=""></option></select>')
                    .appendTo( $(column.header()) )
                    .on( 'change', function () {
                        var val = $.fn.dataTable.util.escapeRegex(
                            $(this).val()
                        );
 
                        column
                            .search( val ? '^'+val+'$' : '', true, false )
                            .draw();
                    } );
 
                column.data().unique().sort().each( function ( d, j ) {
                    var val = $('<div/>').html(d).text();
                    select.append( '<option value="' + val + '">' + val + '</option>' );
                } );
                $( select ).click( function(e) {
                    e.stopPropagation();
              });
            } );
        }
    } );
    // Simulate clearing any search so that we can get the POST quantity fields that DataTables may have removed from the DOM.
    $('button').click( function() {
        var data = table.$('input').serialize();
        const e = $.Event('paste');
        $('[aria-controls="dataTable"]').val('').trigger(e);
        $('select').each( function () { $(this).prop("selectedIndex", 0).trigger("change")});
        return true;
    } );
} );
