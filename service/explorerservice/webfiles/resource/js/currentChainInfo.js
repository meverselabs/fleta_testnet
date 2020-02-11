var CurrentChainInfoAjax={
    reload : function() {
        $.ajax({
            url : "/data/currentChainInfo.data",
            dataType : 'json',
            success : function (data) {
                $("#total_formulators").html(numberWithCommas(data.foumulators))
                $("#total_blocks").html(numberWithCommas(data.blocks))
                $("#total_transactions").html(numberWithCommas(data.transactions))
            }
        })
    },
    init:function(recursive){
        CurrentChainInfoAjax.reload()
        if (recursive !== false) {
            setInterval( function () {
                CurrentChainInfoAjax.reload();
            }, 3000 );
        }
    }
};
