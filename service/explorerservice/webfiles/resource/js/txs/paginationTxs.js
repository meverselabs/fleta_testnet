var paginationTxsAjax={
    reload:function(){
        paginationTxsAjax.obj.ajax.reload()
    },
    blockState:{
        1:{title:"Success",class:" m-badge--success"},
        2:{title:"Pending",class:"m-badge--brand"},
        3:{title:"Delivered",class:" m-badge--metal"},
        4:{title:"Canceled",class:" m-badge--primary"},
        5:{title:"Info",class:" m-badge--info"},
        6:{title:"Danger",class:" m-badge--danger"},
        7:{title:"Warning",class:" m-badge--warning"}
    },
    obj : null,
    init:function(){
        paginationTxsAjax.obj = $("#fleta_pagination_blocks").DataTable({
            responsive:!0,
            searchDelay:500,
            processing:!0,
            serverSide:!0,
            searching: false,
            info: false,
            ordering: false,
            ajax:"/data/paginationTxs.data",
            columns:[{data:"TxHash"},{data:"BlockHash"},{data:"ChainID"},{data:"Time"},{data:"TxType"}],
            columnDefs:[
                {
                    targets:0,
                    render:function(a,e,t,n){
                        return '<a href="/transactionDetail/?hash='+a+'"><span title="'+a+'" class="blockHashSpan">'+a+'</span></a>'
                    }
                },
                {
                    targets:1,
                    render:function(a,e,t,n){
                        return '<a href="/blockDetail/?hash='+a+'"><span title="'+a+'" class="blockHashSpan">'+a+'</span></a>'
                    }
                },
                {
                    targets:3,
                    render:function(a,e,t,n){
                        var d = new Date(a/1000000)
                        a = formatDate(d, "yyyy-MM-dd hh:mm:ss")

                        var texts = a.split(" ")
                        if (texts.length > 1) {
                            return '<span title="'+a+'">'+texts[1]+'</span>'
                        }
                        return '<span title="'+a+'">'+a+'</span>'
                    }
                },
                {
                    targets:4,
                    render:function(a,e,t,n){
                        return '<span class="m-badge m-badge--wide" style="background-color:'+FletaColor(a)+' !important; color: #eee;">'+a+"</span>"
                    }
                }
            ]
        })

    }
};



