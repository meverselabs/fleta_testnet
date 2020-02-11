var DashboardNodesChart={
    chart : new lineChart(),
    reload : function() {
        DashboardNodesChart.chart.draw()
    },
    init : function (recursive) {
        DashboardNodesChart.chart.target = "fleta-formulators";
        DashboardNodesChart.chart.color = "#716aca";
        DashboardNodesChart.chart.dataUrl = "/data/formulators.data";
        DashboardNodesChart.chart.tooltipPrefix = "Nodes : ";
        DashboardNodesChart.reload();
        if (recursive !== false) {
            setInterval( function () {
                DashboardNodesChart.reload();
            }, 3000 );
        }
    }
}

var DashboardTransactionsChart={
    chart : new lineChart(),
    reload : function() {
        DashboardTransactionsChart.chart.draw()
    },
    init : function (recursive) {
        DashboardTransactionsChart.chart.target = "fleta-Transactions";
        DashboardTransactionsChart.chart.color = "#fd4004";
        DashboardTransactionsChart.chart.dataUrl = "/data/transactions.data";
        DashboardTransactionsChart.chart.tooltipPrefix = "Txs : ";
        // DashboardTransactionsChart.chart.m = {top: 50, right: 20, bottom: 20, left: 50};
        // DashboardTransactionsChart.chart.height = 280;
        DashboardTransactionsChart.reload();
        if (recursive !== false) {
            setInterval( function () {
                DashboardTransactionsChart.reload();
            }, 3000 );
        }
    }
}

// var DashboardTransactionsChart={
//     chart : new lineChart(),
//     reload : function() {
//         DashboardTransactionsChart.chart.draw()
//     },
//     init : function (recursive) {
//         DashboardTransactionsChart.chart.target = "fleta-Transactions";
//         DashboardTransactionsChart.chart.color = "#34bfa3";
//         DashboardTransactionsChart.chart.dataUrl = "/data/transactions.data";
//         DashboardTransactionsChart.chart.tooltipPrefix = "Txs : ";
//         DashboardTransactionsChart.reload();
//         if (recursive !== false) {
//             setInterval( function () {
//                 DashboardTransactionsChart.reload();
//             }, 3000 );
//         }
//     }
// }

