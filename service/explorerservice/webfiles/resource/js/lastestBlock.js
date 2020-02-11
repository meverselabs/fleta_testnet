var LastestBlocksAjax={
    lestestBlockTemplate: `
<tr class="{oddeven}">
    <td><a href="/blockDetail?height={Block Height}" title="{Block Hash}" target="_BLANK">{Block Height}</a></td>
    <td><a href="/blockDetail?hash={Block Hash}" title="{Block Hash}"target="_BLANK">{Block Hash}</a></td>
    <td class="{hidelv2}"><span title="{Time}">{Shot Time}</span></td>
    <td class="{hidelv1}"><span class="badge badge-{Status}">{Status}</span></td>
    <td class="{hidelv2}">{Txs}</td>
</tr>
    `,
    observersTemplate: `
<tr class="{oddeven}">
    <td><a href="/blockDetail?height={Block Height}" title="{Block Hash}" target="_BLANK">{Block Height}</a></td>
    <td>{Formulator}</td>
    <td>{OB1}</td>
    <td>{OB2}</td>
    <td>{OB3}</td>
    <td>{OB4}</td>
    <td>{OB5}</td>
</tr>
    `,
    formulratorTemplate: `
<tr class="{oddeven}">
    <td><a href="/blockDetail?height={Block Height}" title="{Block Hash}" target="_BLANK">{Block Height}</a></td>
    <td>{Formulator}</td>
    <td>{BlockCount}</td>
</tr>
    `,
    observersPubKeys :{},
    blockState:{
        1:"Success",
        2:"Fail",
    },
    lastestBlocks : function (data, tWidth, i) {
        var t = LastestBlocksAjax.lestestBlockTemplate+""
        t = t.replace(/{oddeven}/g, (i%2==0)?"odd":"even")

        var texts = data.Time.split(" ")
        if (texts.length > 1) {
            data["Shot Time"] = texts[1]
        } else {
            data["Shot Time"] = data.Time
        }
        if (tWidth > 710) {
            t = t.replace(/{Shot Time}/g, "{Time}")
        }

        if (tWidth <= 710) {
            t = t.replace(/{hidelv1}/g, "hide")
        }
        if (tWidth <= 400) {
            t = t.replace(/{hidelv2}/g, "hide")
        }
        data.Status = void 0===LastestBlocksAjax.blockState[data.Status]?"Fail":LastestBlocksAjax.blockState[data.Status]

        for (var k in data) {
            if (data.hasOwnProperty(k)) {
                var v = data[k]
                t = t.replace(new RegExp("{"+k+"}", 'g'), v)
            }
        }

        return t
    },
    Observers:function(data, i){
        var t = LastestBlocksAjax.observersTemplate+""
        t = t.replace(/{oddeven}/g, (i%2==0)?"odd":"even")
        for (var j = 0 ; j < 5 ; j++) {
            data["OB"+(j+1)] = "-"
        }
        for (var s in data.Signs) {
            var sig = new Buffer(data.Signs[s], "hex");
            ph = pkToPHash(secp256k1.recoverPubKey(data.Msg, sig))
            if (void 0 === LastestBlocksAjax.observersPubKeys[ph]) {
                LastestBlocksAjax.observersPubKeys[ph] = Object.keys(LastestBlocksAjax.observersPubKeys).length
            }
            data["OB"+(LastestBlocksAjax.observersPubKeys[ph]+1)] = "O"
        }
        for (var k in data) {
            if (data.hasOwnProperty(k)) {
                var v = data[k]
                t = t.replace(new RegExp("{"+k+"}", 'g'), v)
            }
        }
        return t
    },
    Formulrator:function(data, i){
        var t = LastestBlocksAjax.formulratorTemplate+""
        t = t.replace(/{oddeven}/g, (i%2==0)?"odd":"even")
        for (var k in data) {
            if (data.hasOwnProperty(k)) {
                var v = data[k]
                t = t.replace(new RegExp("{"+k+"}", 'g'), v)
            }
        }
        return t;
        
    },
    reload:function(){
        $.ajax({
            url : "/data/lastestBlocks.data",
            dataType : 'json',
            success : function (d) {
                var data = d.aaData;
                var tbody = $("#fleta_blocks tbody");
                var ths = $("#fleta_blocks thead th");
                tbody.empty();
                
                ths.removeAttr("style")
                if (tbody.width() <= 710) {
                    ths.eq(3).hide()
                }
                if (tbody.width() <= 400) {
                    ths.eq(2).hide()
                    ths.eq(4).hide()
                }

                var otbody = $("#fletaObservers tbody");
                otbody.empty();

                var ftbody = $("#fletaFormulrator tbody");
                ftbody.empty();

                for (var i = 0 ; i < data.length ; i++) {
                    tbody.append(LastestBlocksAjax.lastestBlocks(data[i], tbody.width(), i))
                    otbody.append(LastestBlocksAjax.Observers(data[i], i))
                    ftbody.append(LastestBlocksAjax.Formulrator(data[i], i))
                }
            }
        })
    },
    init:function(recursive){
        LastestBlocksAjax.reload()
        if (recursive !== false) {
            setInterval( function () {
                LastestBlocksAjax.reload();
            }, 3000 );
        }
    }
};



