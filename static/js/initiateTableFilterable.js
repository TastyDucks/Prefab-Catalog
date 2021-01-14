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
} );
