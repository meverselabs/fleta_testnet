var ChainInfoTableAjax={
    reload:function(){
        ChainInfoTableAjax.obj.ajax.reload()
    },
    obj : null,
    init:function(recursive){
        ChainInfoTableAjax.obj = $("#fleta_chain_info").DataTable({
            responsive:!0,
            searchDelay:500,
            processing:!0,
            serverSide:!0,
            paging: false,
            searching: false,
            info: false,
            ordering: false,
            ajax:"/data/chainInfoTable.data",
            columns:[{data:"구분"},{data:"블록 크기"},{data:"블록 전송 시간"},{data:"블록 연결 시간"}],
            columnDefs:[
                {
                    targets:0,
                    render:function(a,e,t,n){
                        return a
                    }
                },
            ]
        })

        if (recursive !== false) {
            setInterval( function () {
                ChainInfoTableAjax.reload();
            }, 3000 );
        }
    }
};



