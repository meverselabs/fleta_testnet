/**
* LICENSE#
* Released under the GNU General Public License, version 3.
* 
* https://bl.ocks.org/mbostock/1256572
*/
function perBlock () {}
perBlock.prototype.symbols = null
perBlock.prototype.svg = null
perBlock.prototype.delay = 500
perBlock.prototype.duration = 1500
perBlock.prototype.line = null
perBlock.prototype.axis = null
perBlock.prototype.area = null
perBlock.prototype.x = null
perBlock.prototype.y = null

perBlock.prototype.Start = function (symbols) {
    var This = perBlock.prototype

    var $ttpb = $("#transaction_type_per_block")
    $ttpb.html("")

    var m = [20, 20, 30, 20];
    This.w = $ttpb.width() - m[1] - m[3],
    This.h = 200 - m[0] - m[2];
    
    This.svg = d3.select("#transaction_type_per_block").append("svg")
        .attr("width", This.w + m[1] + m[3])
        .attr("height", This.h + m[0] + m[2])
        .append("g")
        .attr("transform", "translate(" + m[3] + "," + m[0] + ")");
    
    This.stocks;

    // A line generator, for the dark stroke.
    This.line = d3.svg.line()
        .interpolate("basis")
        .x(function(d) { return This.x(d.blockTime); })
        .y(function(d) { return This.y(d.txCount); });

    // A line generator, for the dark stroke.
    This.axis = d3.svg.line()
        .interpolate("basis")
        .x(function(d) { return This.x(d.blockTime); })
        .y(This.h);
    
    // A area generator, for the dark stroke.
    This.area = d3.svg.area()
        .interpolate("basis")
        .x(function(d) { return This.x(d.blockTime); })
        .y1(function(d) { return This.y(d.txCount); });
    
    
    function drawChart() {
        This.svg.selectAll("g")
        .data(This.symbols)
        .enter().append("g")
        .attr("class", "symbol");

        setTimeout(This.lines, This.duration);
    }
    
    if (typeof symbols !== "undefined") {
        This.symbols = symbols
        drawChart()
        return;
    }

    d3.json("./data/typesPerBlock.data", function(data) {
    // d3.csv("./data2/data.csv", function(data) {
        // Nest stock values by symbol.
        This.symbols = d3.nest()
            .key(function(d) { return d.symbol; })
            .sortValues(function(a,b) {
                return ((+a.blockTime < +b.blockTime) ? -1 : 1);
            })
            .entries(This.stocks = data);

        // Parse dates and numbers. We assume values are sorted by date.
        // Also compute the maximum txCount per symbol, needed for the y-domain.
        This.symbols.forEach(function(s) {
            s.values.forEach(function(d) { d.txCount = +d.txCount; });
            s.maxTxCount = d3.max(s.values, function(d) { return d.txCount; });
            s.sumTxCount = d3.sum(s.values, function(d) { return d.txCount; });
        });

        // Sort by maximum txCount, descending.
        This.symbols.sort(function(a, b) { return b.maxTxCount - a.maxTxCount; });

        drawChart()

    });
}

perBlock.prototype.lines = function() {
    var This = perBlock.prototype

    This.x = d3.time.scale().range([0, This.w - 60]);
    This.y = d3.scale.linear().range([This.h / 4 - 20, 0]);

    // Compute the minimum and maximum date across symbols.
    This.x.domain([
        d3.min(This.symbols, function(d) { return d.values[0].blockTime; }),
        d3.max(This.symbols, function(d) { return d.values[d.values.length - 1].blockTime; })
    ]);
    
    var g = This.svg.selectAll(".symbol")
    .attr("transform", function(d, i) { return "translate(0," + (i * This.h / 4 + 10) + ")"; });
    
    g.each(function(d) {
        var e = d3.select(this);
        
        e.append("path")
            .attr("class", "line");
        
        e.append("circle")
            .attr("r", 5)
            .style("fill", function(d) { return FletaColor(d.key); })
            .style("stroke", "#000")
            .style("stroke-width", "2px");
        
        e.append("text")
            .attr("x", 12)
            .attr("dy", ".31em")
            .text(d.key);
    });
    
    function draw(k) {
        g.each(function(d) {
            var e = d3.select(this);
            This.y.domain([0, d.maxTxCount]);
            
            e.select("path")
                .attr("d", function(d) { return This.line(d.values.slice(0, k + 1)); });
            
            e.selectAll("circle, text")
                .data(function(d) { return [d.values[k], d.values[k]]; })
                .attr("transform", function(d) { return "translate(" + This.x(d.blockTime) + "," + This.y(d.txCount) + ")"; });
        });
    }
    
    var k = 1, n = This.symbols[0].values.length;
    d3.timer(function() {
        draw(k);
        if ((k += 2) >= n - 1) {
            draw(n - 1);
            setTimeout(This.horizons, 500);
            return true;
        }
    });
}

perBlock.prototype.horizons = function() {
    var This = perBlock.prototype

    This.svg.insert("defs", ".symbol")
        .append("clipPath")
        .attr("id", "clip")
        .append("rect")
        .attr("width", This.w)
        .attr("height", This.h / 4 - 20);
    
    var color = d3.scale.ordinal()
    .range(["#c6dbef", "#9ecae1", "#6baed6"]);
    
    var g = This.svg.selectAll(".symbol")
        .attr("clip-path", "url(#clip)");
    
    This.area
        .y0(This.h / 4 - 20);
    
    g.select("circle").transition()
        .duration(This.duration)
        .attr("transform", function(d) { return "translate(" + (This.w - 60) + "," + (-This.h / 4) + ")"; })
        .remove();
    
    g.select("text").transition()
        .duration(This.duration)
        .attr("transform", function(d) { return "translate(" + (This.w - 60) + "," + (This.h / 4 - 20) + ")"; })
        .attr("dy", "0em");
    
    g.each(function(d) {
        This.y.domain([0, d.maxTxCount]);
    
        d3.select(this).selectAll(".area")
            .data(d3.range(3))
        .enter().insert("path", ".line")
            .attr("class", "area")
            .attr("transform", function(d) { return "translate(0," + (d * (This.h / 4 - 20)) + ")"; })
            .attr("d", This.area(d.values))
            .style("fill", function(d, i) { return color(i); })
            .style("fill-opacity", 1e-6);
        
        This.y.domain([0, d.maxTxCount / 3]);
        
        d3.select(this).selectAll(".line").transition()
            .duration(This.duration)
            .attr("d", This.line(d.values))
            .style("stroke-opacity", 1e-6);
        
        d3.select(this).selectAll(".area").transition()
            .duration(This.duration)
            .style("fill-opacity", 1)
            .attr("d", This.area(d.values))
            .each("end", function() { d3.select(this).style("fill-opacity", null); });
    });
    
    setTimeout(This.areas, This.duration + This.delay);
}

perBlock.prototype.areas = function() {
    var This = perBlock.prototype

    var g = This.svg.selectAll(".symbol");
    
    This.axis
        .y(This.h / 4 - 21);
    
    g.select(".line")
        .attr("d", function(d) { return This.axis(d.values); });
    
    g.each(function(d) {
        This.y.domain([0, d.maxTxCount]);
        
        d3.select(this).select(".line").transition()
            .duration(This.duration)
            .style("stroke-opacity", 1)
            .each("end", function() { d3.select(this).style("stroke-opacity", null); });
        
        d3.select(this).selectAll(".area")
            .filter(function(d, i) { return i; })
        .transition()
            .duration(This.duration)
            .style("fill-opacity", 1e-6)
            .attr("d", This.area(d.values))
            .remove();
        
        d3.select(this).selectAll(".area")
            .filter(function(d, i) { return !i; })
        .transition()
            .duration(This.duration)
            .style("fill", FletaColor(d.key))
            .attr("d", This.area(d.values));
    });
    
    This.svg.select("defs").transition()
        .duration(This.duration)
        .remove();
    
    g.transition()
        .duration(This.duration)
        .each("end", function() { d3.select(this).attr("clip-path", null); });
    
    setTimeout(This.stackedArea, This.duration + This.delay);
}

perBlock.prototype.stackedArea = function() {
    var This = perBlock.prototype

    var stack = d3.layout.stack()
    .values(function(d) { return d.values; })
    .x(function(d) { return d.blockTime; })
    .y(function(d) { return d.txCount; })
    .out(function(d, y0, y) { d.txCount0 = y0; })
    .order("reverse");
    
    stack(This.symbols);
    
    This.y
        .domain([0, d3.max(This.symbols[0].values.map(function(d) { return d.txCount + d.txCount0; }))])
        .range([This.h, 0]);
    
    This.line
        .y(function(d) { return This.y(d.txCount0); });
    
    This.area
        .y0(function(d) { return This.y(d.txCount0); })
        .y1(function(d) { return This.y(d.txCount0 + d.txCount); });
    
    var t = This.svg.selectAll(".symbol").transition()
        .duration(This.duration)
        .attr("transform", "translate(0,0)")
        .each("end", function() { d3.select(this).attr("transform", null); });
    
    t.select("path.area")
        .attr("d", function(d) { return This.area(d.values); });
    
    t.select("path.line")
        .style("stroke-opacity", function(d, i) { return i < 3 ? 1e-6 : 1; })
        .attr("d", function(d) { return This.line(d.values); });
    
    t.select("text")
        .attr("transform", function(d) { d = d.values[d.values.length - 1]; return "translate(" + (This.w - 60) + "," + This.y(d.txCount / 2 + d.txCount0) + ")"; });
    
    setTimeout(This.overlappingArea, This.duration + This.delay);
}

perBlock.prototype.overlappingArea = function() {
    var This = perBlock.prototype

    var g = This.svg.selectAll(".symbol");
    
    This.line
        .y(function(d) { return This.y(d.txCount0 + d.txCount); });
    
    g.select(".line")
        .attr("d", function(d) { return This.line(d.values); });
    
    This.y
        .domain([0, d3.max(This.symbols.map(function(d) { return d.maxTxCount; }))])
        .range([This.h, 0]);
    
    This.area
        .y0(This.h)
        .y1(function(d) { return This.y(d.txCount); });
    
    This.line
        .y(function(d) { return This.y(d.txCount); });
    
    var t = g.transition()
        .duration(This.duration);
    
    t.select(".line")
        .style("stroke-opacity", 1)
        .attr("d", function(d) { return This.line(d.values); });
    
    t.select(".area")
        .style("fill-opacity", .5)
        .attr("d", function(d) { return This.area(d.values); });
    
    t.select("text")
        .attr("dy", ".31em")
        .attr("transform", function(d) { d = d.values[d.values.length - 1]; return "translate(" + (This.w - 60) + "," + This.y(d.txCount) + ")"; });
    
    This.svg.append("line")
        .attr("class", "line")
        .attr("x1", 0)
        .attr("x2", This.w - 60)
        .attr("y1", This.h)
        .attr("y2", This.h)
        .style("stroke-opacity", 1e-6)
    .transition()
        .duration(This.duration)
        .style("stroke-opacity", 1);
    
    setTimeout(This.groupedBar, This.duration + This.delay);
}

perBlock.prototype.groupedBar = function() {
    var This = perBlock.prototype

    This.x = d3.scale.ordinal()
    .domain(This.symbols[0].values.map(function(d) { return d.blockTime; }))
    .rangeBands([0, This.w - 60], .1);
    
    var x1 = d3.scale.ordinal()
        .domain(This.symbols.map(function(d) { return d.key; }))
        .rangeBands([0, This.x.rangeBand()]);
    
    var g = This.svg.selectAll(".symbol");
    
    var t = g.transition()
        .duration(This.duration);
    
    t.select(".line")
        .style("stroke-opacity", 1e-6)
        .remove();
    
    t.select(".area")
        .style("fill-opacity", 1e-6)
        .remove();
    
    g.each(function(p, j) {
    d3.select(this).selectAll("rect")
        .data(function(d) { return d.values; })
    .enter().append("rect")
        .attr("x", function(d) { return This.x(d.blockTime) + x1(p.key); })
        .attr("y", function(d) { return This.y(d.txCount); })
        .attr("width", x1.rangeBand())
        .attr("height", function(d) { return This.h - This.y(d.txCount); })
        .style("fill", FletaColor(p.key))
        .style("fill-opacity", 1e-6)
    .transition()
        .duration(This.duration)
        .style("fill-opacity", 1);
    });
    
    setTimeout(This.stackedBar, This.duration + This.delay);
}

perBlock.prototype.stackedBar = function() {
    var This = perBlock.prototype

    This.x.rangeRoundBands([0, This.w - 60], .1);
    
    var stack = d3.layout.stack()
    .values(function(d) { return d.values; })
    .x(function(d) { return d.blockTime; })
    .y(function(d) { return d.txCount; })
    .out(function(d, y0, y) { d.txCount0 = y0; })
    .order("reverse");
    
    var g = This.svg.selectAll(".symbol");
    
    stack(This.symbols);
    
    This.y
        .domain([0, d3.max(This.symbols[0].values.map(function(d) { return d.txCount + d.txCount0; }))])
        .range([This.h, 0]);
    
    var t = g.transition()
        .duration(This.duration / 2);
    
    t.select("text")
        .delay(This.symbols[0].values.length * 10)
    .attr("transform", function(d) { d = d.values[d.values.length - 1]; return "translate(" + (This.w - 60) + "," + This.y(d.txCount / 2 + d.txCount0) + ")"; });

    t.selectAll("rect")
        .delay(function(d, i) { return i * 10; })
        .attr("y", function(d) { return This.y(d.txCount0 + d.txCount); })
        .attr("height", function(d) { return This.h - This.y(d.txCount); })
        .each("end", function() {
            d3.select(this)
                .style("stroke", "#fff")
                .style("stroke-opacity", 1e-6)
            .transition()
                .duration(This.duration / 2)
                .attr("x", function(d) { return This.x(d.blockTime); })
                .attr("width", This.x.rangeBand())
                .style("stroke-opacity", 1);
        });
    
    setTimeout(This.transposeBar, This.duration + This.symbols[0].values.length * 10 + This.delay);
}

perBlock.prototype.transposeBar = function() {
    var This= perBlock.prototype
    This.x
        .domain(This.symbols.map(function(d) { return d.key; }))
        .rangeRoundBands([0, This.w], .2);
    
    This.y
        .domain([0, d3.max(This.symbols.map(function(d) { return d3.sum(d.values.map(function(d) { return d.txCount; })); }))]);
    
    var stack = d3.layout.stack()
        .x(function(d, i) { return i; })
        .y(function(d) { return d.txCount; })
        .out(function(d, y0, y) { d.txCount0 = y0; });
    
    stack(d3.zip.apply(null, This.symbols.map(function(d) { return d.values; }))); // transpose!
    
    var g = This.svg.selectAll(".symbol");
    
    var t = g.transition()
        .duration(This.duration / 2);
    
    t.selectAll("rect")
        .delay(function(d, i) { return i * 10; })
        .attr("y", function(d) { return This.y(d.txCount0 + d.txCount) - 1; })
        .attr("height", function(d) { return This.h - This.y(d.txCount) + 1; })
        .attr("x", function(d) { return This.x(d.symbol); })
        .attr("width", This.x.rangeBand())
        .style("stroke-opacity", 1e-6);
    
    t.select("text")
        .attr("x", 0)
        .attr("transform", function(d) { return "translate(" + (This.x(d.key) + This.x.rangeBand() / 2) + "," + This.h + ")"; })
        .attr("dy", "1.31em")
        .each("end", function() { d3.select(this).attr("x", null).attr("text-anchor", "middle"); });
    
    This.svg.select("line").transition()
        .duration(This.duration)
        .attr("x2", This.w);
    
    setTimeout(This.donut,  This.duration / 2 + This.symbols[0].values.length * 10 + This.delay);
}

perBlock.prototype.donut = function() {
    var This = perBlock.prototype
    var g = This.svg.selectAll(".symbol");
    
    g.selectAll("rect").remove();
    
    var pie = d3.layout.pie()
    .value(function(d) { return d.sumTxCount; });
    
    var arc = d3.svg.arc();
    
    g.append("path")
        .style("fill", function(d) { return FletaColor(d.key); })
        .data(function() { return pie(This.symbols); })
        .transition()
        .duration(This.duration)
        .tween("arc", arcTween);
    
    g.select("text").transition()
        .duration(This.duration)
        .attr("dy", ".31em");
    
    This.svg.select("line").transition()
        .duration(This.duration)
        .attr("y1", 2 * This.h)
        .attr("y2", 2 * This.h)
        .remove();
    
    function arcTween(d) {
        var path = d3.select(this),
            text = d3.select(this.parentNode.appendChild(this.previousSibling)),
            x0 = This.x(d.data.key),
            y0 = This.h - This.y(d.data.sumTxCount);
        
        return function(t) {
            var r = This.h / 2 / Math.min(1, t + 1e-3),
                a = Math.cos(t * Math.PI / 2),
                xx = (-r + (a) * (x0 + This.x.rangeBand()) + (1 - a) * (This.w + This.h) / 2),
                yy = ((a) * This.h + (1 - a) * This.h / 2),
                f = {
                    innerRadius: r - This.x.rangeBand() / (2 - a),
                    outerRadius: r,
                    startAngle: a * (Math.PI / 2 - y0 / r) + (1 - a) * d.startAngle,
                    endAngle: a * (Math.PI / 2) + (1 - a) * d.endAngle
                };
            
            path.attr("transform", "translate(" + xx + "," + yy + ")");
            path.attr("d", arc(f));
            text.attr("transform", "translate(" + arc.centroid(f) + ")translate(" + xx + "," + yy + ")rotate(" + ((f.startAngle + f.endAngle) / 2 + 3 * Math.PI / 2) * 180 / Math.PI + ")");
        };
    }
    
    setTimeout(This.donutExplode, This.duration + This.delay);
}

perBlock.prototype.donutExplode = function() {
    var This = perBlock.prototype

    var r0a = This.h / 2 - This.x.rangeBand() / 2,
    r1a = This.h / 2,
    r0b = 2 * This.h - This.x.rangeBand() / 2,
    r1b = 2 * This.h,
    arc = d3.svg.arc();
    
    This.svg.selectAll(".symbol path")
    .each(transitionExplode);
    
    function transitionExplode(d, i) {
        d.innerRadius = r0a;
        d.outerRadius = r1a;
        d3.select(this).transition()
            .duration(This.duration / 2)
            .tween("arc", tweenArc({
                innerRadius: r0b,
                outerRadius: r1b
            }));
    }
    
    function tweenArc(b) {
        return function(a) {
            var path = d3.select(this),
                text = d3.select(this.nextSibling),
                i = d3.interpolate(a, b);
            for (var key in b) a[key] = b[key]; // update data
            return function(t) {
                var a = i(t);
                path.attr("d", arc(a));
                text.attr("transform", "translate(" + arc.centroid(a) + ")translate(" + This.w / 2 + "," + This.h / 2 +")rotate(" + ((a.startAngle + a.endAngle) / 2 + 3 * Math.PI / 2) * 180 / Math.PI + ")");
            };
        }
    }
}

var TransactionTypePerBlockAjax={
    obj : new perBlock(),
    reload: function () {
        TransactionTypePerBlockAjax.obj.Start()
        setInterval(function() {
            TransactionTypePerBlockAjax.obj.svg.selectAll("*").remove();
            TransactionTypePerBlockAjax.obj.Start()
        }, (TransactionTypePerBlockAjax.obj.duration*11) + (TransactionTypePerBlockAjax.obj.delay * 10));
    },
    init: function () {
        TransactionTypePerBlockAjax.reload()
    }
}
