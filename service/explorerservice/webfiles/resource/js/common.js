var FletaColor = d3.scale.ordinal()
    .domain(["Transfer", "Withdraw", "Burn", "CreateAccount", "CreateMultiSigAccount","Assign","Deposit","OpenAccount","CreateFormulation","RevokeFormulation","SolidityCreateContract","SolidityCallContract"])
    .range([
        "#36a3f7","#ffb822","#716aca","#34bfa3","#5867dd","#00c5dc","#716aca","#f4516c","#bf3482","#34bf7d","#98bf34","#bf7934"
    ]);

function lineChart () {}
lineChart.prototype.m = {top: 5, right: 0, bottom: 20, left: 50};
lineChart.prototype.height = 200;
lineChart.prototype.draw = function () {
    var This = this
    d3.json(This.dataUrl, d => {
        // d[100].count = 10000
        if (typeof d == "undefined" || d == null) {
            return
        }
        var $ff = $("#"+This.target)

        $ff.html("")
        var $fft = $('#'+This.target+'-tooltip')
        if ($fft.length == 0) {
            $("<div id='"+This.target+'-tooltip'+"' class='fleta-chart-tooltip'>").insertAfter($ff)
        }

        This.w = $ff.width() - This.m.right - This.m.left,
        This.h = This.height - This.m.top - This.m.bottom;

        This.svg = d3.select("#"+This.target).append("svg")
        .attr("width", This.w + This.m.right + This.m.left)
        .attr("height", This.h + This.m.top + This.m.bottom)

        if (typeof d[0] != "undefined") {
            This.countMax = d3.max(d, d => d.count);
            This.countMin = d3.min(d, d => {
                d.time = parseInt(d.time/1000000000)
                // d.time = formatDate(new Date(d.time/1000000), "yyyy-MM-dd hh:mm:ss")
                return d.count
            });

            This.timeMin = d3.min(d, d => d.time);
            This.timeMax = d3.max(d, d => {
                d.time -= This.timeMin
                return d.time
            });

            var yScale = d3.scale.linear().domain([0,This.countMax]).range([This.h+This.m.top, This.m.top])
            var yAxis = d3.svg.axis().scale(yScale).orient("left")
            .innerTickSize(-This.w)
            .outerTickSize(0)
            .tickPadding(5)

            var xScale = d3.scale.linear().domain([This.timeMin,This.timeMin+This.timeMax]).range([This.m.left,This.w+This.m.left])
            var xAxis = d3.svg.axis().scale(xScale).orient("bottom").tickFormat(d => {
                var d = new Date(d*1000)
                return formatDate(d, "HH:mm:ss");
            })
            .innerTickSize(-This.h)
            .outerTickSize(0)
            .tickPadding(10)

            This.yAxis = This.svg.append("g")
                .attr("class", "y axis")
                .attr("transform", "translate(" + This.m.left + ",0)")
                .call(yAxis);

            This.xAxis = This.svg.append("g")
                .attr("class", "x axis")
                .attr("transform", "translate(0," + (This.h + This.m.top) + ")")
                .call(xAxis)

            This.svg
                .selectAll(".axis line, .axis path")
                .attr("fill", "none")
                .attr("stroke", "black")
            This.svg
                .selectAll(".tick line")
                .attr("opacity", "0.2")

            This.chart = This.svg
                .append("g")
                .attr("transform", "translate(" + This.m.left + "," + This.m.top + ")");

            This.tooltip = d3.select('#'+This.target+'-tooltip');
            This.tooltipLine = This.chart.append('line');

            This.x = d3.scale.linear().domain([0, This.timeMax]).range([0, This.w]);
            This.y = d3.scale.linear().domain([0, This.countMax]).range([This.h, 0]);

            This.line = d3.svg.line().x(d => This.x(d.time)).y(d => This.y(d.count));
		}

		if (d.length > 0) {
			for (var i = 1 ; i < d.length ; i++) {
				if (d[i-1].time < d[i].time-1) {
					d.splice(i, 0, {
						time : parseInt((d[i-1].time+d[i].time)/2),
						count : (d[i-1].count+d[i].count)/2,
						empty : true,
					})
					i++
					console.log("noe")
				}
			}
		}

		This.states = d;

        This.chart
            .append("g")
            .append('path')
            .attr('fill', 'none')
            .attr('stroke', This.color)
            .attr('stroke-width', 2)
            .datum(This.states)
            .attr('d', This.line);

        This.chart
            .append("g").selectAll("circle")
            .data(This.states).enter()
            .append('circle')
            .attr('r', '4')
            .attr('cx', d=>This.x(d.time))
            .attr('cy', d=>This.y(d.count))
            .attr('fill', d=>d.empty?"#4004fd":This.color)

        This.chart
            .append("g").selectAll("circle")
            .data(This.states).enter()
            .append('circle')
            .attr('r', '2')
            .attr('cx', d=>This.x(d.time))
            .attr('cy', d=>This.y(d.count))
            .attr('fill', "#fff")
            // .attr('opacity', 0.5)

        This.tipBox = This.chart.append('rect')
            .attr('width', This.w)
            .attr('height', This.h)
            .attr('opacity', 0)
            .on('mousemove', This.drawTooltip(This))
            .on('mouseout', This.removeTooltip(This));
    })

}

lineChart.prototype.removeTooltip = function(This) {
    return () => {
        if (This.tooltip) This.tooltip.style('display', 'none');
        if (This.tooltipLine) This.tooltipLine.attr('stroke', 'none');
    }
}

lineChart.prototype.drawTooltip = function (This) {
    return () => {
        This.tooltip.style('display', 'block');
        const time = Math.floor(This.x.invert(d3.mouse(This.tipBox.node())[0])+0.5);

        This.states.sort((a, b) => {
            return b.count - a.count;
        })

        This.tooltipLine.attr('stroke', 'black')
            .attr('x1', () => {
                return This.x(time)
            })
            .attr('x2', () => {
                return This.x(time)
            })
            .attr('y1', 0)
            .attr('y2', This.h);

        var coordinates = [0, 0];
        coordinates = d3.mouse(This.svg.node());
        var x = coordinates[0];
        var y = coordinates[1];

        var d = new Date((This.timeMin*1000) + (time*1000))
        This.tooltip
            .style("left", (x + 20 )+"px")
            .style("top", (y + 100 )+"px")
            .html(formatDate(d, "yyyy-MM-dd<br/>HH:mm:ss"))
            .append('div')
            .style('color', This.color)
            .html(() => {
                return This.tooltipPrefix + This.states.find(h => h.time == time).count
            });
    }
}

function formatDate(date, format, utc) {
    var MMMM = ["\x00", "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
    var MMM = ["\x01", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    var dddd = ["\x02", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
    var ddd = ["\x03", "Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

    function ii(i, len) {
        var s = i + "";
        len = len || 2;
        while (s.length < len) s = "0" + s;
        return s;
    }

    var y = utc ? date.getUTCFullYear() : date.getFullYear();
    format = format.replace(/(^|[^\\])yyyy+/g, "$1" + y);
    format = format.replace(/(^|[^\\])yy/g, "$1" + y.toString().substr(2, 2));
    format = format.replace(/(^|[^\\])y/g, "$1" + y);

    var M = (utc ? date.getUTCMonth() : date.getMonth()) + 1;
    format = format.replace(/(^|[^\\])MMMM+/g, "$1" + MMMM[0]);
    format = format.replace(/(^|[^\\])MMM/g, "$1" + MMM[0]);
    format = format.replace(/(^|[^\\])MM/g, "$1" + ii(M));
    format = format.replace(/(^|[^\\])M/g, "$1" + M);

    var d = utc ? date.getUTCDate() : date.getDate();
    format = format.replace(/(^|[^\\])dddd+/g, "$1" + dddd[0]);
    format = format.replace(/(^|[^\\])ddd/g, "$1" + ddd[0]);
    format = format.replace(/(^|[^\\])dd/g, "$1" + ii(d));
    format = format.replace(/(^|[^\\])d/g, "$1" + d);

    var H = utc ? date.getUTCHours() : date.getHours();
    format = format.replace(/(^|[^\\])HH+/g, "$1" + ii(H));
    format = format.replace(/(^|[^\\])H/g, "$1" + H);

    var h = H > 12 ? H - 12 : H == 0 ? 12 : H;
    format = format.replace(/(^|[^\\])hh+/g, "$1" + ii(h));
    format = format.replace(/(^|[^\\])h/g, "$1" + h);

    var m = utc ? date.getUTCMinutes() : date.getMinutes();
    format = format.replace(/(^|[^\\])mm+/g, "$1" + ii(m));
    format = format.replace(/(^|[^\\])m/g, "$1" + m);

    var s = utc ? date.getUTCSeconds() : date.getSeconds();
    format = format.replace(/(^|[^\\])ss+/g, "$1" + ii(s));
    format = format.replace(/(^|[^\\])s/g, "$1" + s);

    var f = utc ? date.getUTCMilliseconds() : date.getMilliseconds();
    format = format.replace(/(^|[^\\])fff+/g, "$1" + ii(f, 3));
    f = Math.round(f / 10);
    format = format.replace(/(^|[^\\])ff/g, "$1" + ii(f));
    f = Math.round(f / 10);
    format = format.replace(/(^|[^\\])f/g, "$1" + f);

    var T = H < 12 ? "AM" : "PM";
    format = format.replace(/(^|[^\\])TT+/g, "$1" + T);
    format = format.replace(/(^|[^\\])T/g, "$1" + T.charAt(0));

    var t = T.toLowerCase();
    format = format.replace(/(^|[^\\])tt+/g, "$1" + t);
    format = format.replace(/(^|[^\\])t/g, "$1" + t.charAt(0));

    var tz = -date.getTimezoneOffset();
    var K = utc || !tz ? "Z" : tz > 0 ? "+" : "-";
    if (!utc) {
        tz = Math.abs(tz);
        var tzHrs = Math.floor(tz / 60);
        var tzMin = tz % 60;
        K += ii(tzHrs) + ":" + ii(tzMin);
    }
    format = format.replace(/(^|[^\\])K/g, "$1" + K);

    var day = (utc ? date.getUTCDay() : date.getDay()) + 1;
    format = format.replace(new RegExp(dddd[0], "g"), dddd[day]);
    format = format.replace(new RegExp(ddd[0], "g"), ddd[day]);

    format = format.replace(new RegExp(MMMM[0], "g"), MMMM[M]);
    format = format.replace(new RegExp(MMM[0], "g"), MMM[M]);

    format = format.replace(/\\(.)/g, "$1");

    return format;
};

function sendNewAccountTx () {
    $.ajax({
        url : "/tx/CreateAccount.tx",
        success : function () {
            // alert("send")
        }
    })
}
function sendBurnTx () {
    $.ajax({
        url : "/tx/Burn.tx",
        success : function () {
            // alert("send")
        }
    })
}
function sendTransfertx () {
    $.ajax({
        url : "/tx/Transfer.tx",
        success : function () {
            // alert("send")
        }
    })
}
function sendWithdrawtx () {
    $.ajax({
        url : "/tx/Withdraw.tx",
        success : function () {
            // alert("send")
        }
    })
}