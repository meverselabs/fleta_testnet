var LastestTransactionsAjax={
    reload : function() {
        $.ajax({
            url : "/data/lastestTransactions.data",
            dataType : 'json',
            success : function (data) {
                var $txs = $("#fleta-lastest-transactions");

                $txs.html("")

                for (var i = 0 ; i < data.length ; i++) {
                    var r = data[i]
                    var $e = LastestTransactionsAjax.render(r.Time,r.TxHash,r.TxType)
                    $txs.append($e)
                }
            }
        })

    },
    render: function (time, hash, type) {
        var $template = $("#transactions-item-template").clone()
        $template.find(".timeline-3_item").addClass(type.replace(/\./gi, ""))
        var now = new Date(time/1000000)
        
        $template.find("#tx-time").html(formatDate(now, "hh:mm")).attr("title", formatDate(now, "yyyy-MM-dd hh:mm:ss")).removeAttr("id")
        $template.find("#tx-hash-atag").attr("href", "/transactionDetail?hash="+hash)
        $template.find("#tx-hash").html(hash).removeAttr("id")
        $template.find("#tx-type").html(type).removeAttr("id")
        return $template.children()
    },
    init:function(recursive){
        LastestTransactionsAjax.reload();

        if (recursive !== false) {
            setInterval( function () {
                LastestTransactionsAjax.reload();
            }, 3000 );
        }
    }
};



