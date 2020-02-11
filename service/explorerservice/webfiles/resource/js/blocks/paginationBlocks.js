var paginationBlocksAjax={
    reload:function(){
        paginationBlocksAjax.obj.ajax.reload()
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
        paginationBlocksAjax.obj = $("#fleta_pagination_blocks").DataTable({
            responsive:!0,
            searchDelay:500,
            processing:!0,
            serverSide:!0,
            searching: false,
            info: false,
            ordering: false,
            ajax:"/data/paginationBlocks.data",
            columns:[{data:"Block Height"},{data:"Block Hash"},{data:"Time"},{data:"Status"},{data:"Txs"}],
            columnDefs:[
                {
                    targets:1,
                    render:function(a,e,t,n){
                        return '<a href="/blockDetail/?hash='+a+'&height='+t["Block Height"]+'"><span title="'+a+'" class="blockHashSpan">'+a+'</span></a>'
                    }
                },
                {
                    targets:2,
                    render:function(a,e,t,n){
                        var texts = a.split(" ")
                        if (texts.length > 1) {
                            return '<span title="'+a+'">'+texts[1]+'</span>'
                        }
                        return '<span title="'+a+'">'+a+'</span>'
                    }
                },
                {
                    targets:3,
                    render:function(a,e,t,n){
                        return void 0===paginationBlocksAjax.blockState[a]?a:'<span class="m-badge '+paginationBlocksAjax.blockState[a].class+' m-badge--wide">'+paginationBlocksAjax.blockState[a].title+"</span>"
                    }
                }
            ]
        })

    }
};



