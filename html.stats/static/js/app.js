
/**
 *  Set up the Chart.js defaults.
 */
/*
Chart.defaults.global = {
	animationEasing: "easeOutQuart",
	responsive: true,
	showTooltips: true,
	tooltipFillColor: "rgba(0,0,0,0.8)",
	tooltipFontFamily: "'Helvetica Neue', 'Helvetica', 'Arial', sans-serif",
	tooltipFontSize: 14,
	tooltipFontStyle: "normal",
	tooltipFontColor: "#fff",
	tooltipTitleFontFamily: "'Helvetica Neue', 'Helvetica', 'Arial', sans-serif",
	tooltipTitleFontSize: 14,
	tooltipTitleFontStyle: "bold",
	tooltipTitleFontColor: "#fff",
	tooltipTemplate: "Games Played<%if (label){%><%=label%>: <%}%><%= value %>",
};
*/

(function() {
    var app = angular.module("LoLStats", ["angles"]);
    
    app.factory("graphData", function($rootScope) {
		var graphData = {};
		
		graphData.labels = [];
		graphData.data = {};
		graphData.activeMetric = "";
		
		graphData.setLabels = function(labels) {
			graphData.labels = labels;
		}
		
		graphData.getLabels = function() {
			return graphData.labels;
		}
		
		graphData.setData = function(metric, data) {
			graphData.data[metric] = data;
		}
		
		graphData.getData = function() {
			return graphData.data[graphData.activeMetric];
		}
		
		graphData.setActiveMetric = function(metric) {
			graphData.activeMetric = metric;
			$rootScope.$broadcast('updateGraph');
		}
		
		return graphData;
	});
    
    /**
	 * The application-level controller for the full app.
	 */
	app.controller("AppController", function($scope, $http, graphData) {
		// Identify whether the requestedSummoner is a known entity.
		$scope.validSummoner = false;
		$scope.metrics = [];
		$scope.dates = [];
		
		$scope.summary_value = function(metric) {
			if (metric == null) {
				return "XX (+X%)"
			}
			
			var latest = "0000";
			var second_latest = "0000";
			for (var date_string in metric) {
				if (date_string > latest) {
					second_latest = latest;
					latest = date_string;
				}
			}
			
			var delta = (metric[latest] / metric[second_latest]) - 1;
			
			if (delta > 0) {
				return metric[latest] + " (+" + Math.round(delta * 100) / 10 + "%)";
			}
			else {
				return metric[latest] + " (" + Math.round(delta * 1000) / 10 + "%)";	
			}
		}
		
		// This should make a request to get the JSON response for the provided
		// summoner.
		$scope.requestSummoner = function() {			
			$http.get("static/sample.json").success(function(data) {		
			//$http.get("summoners/" + $scope.requestedSummoner).success(function(data) {
				$scope.validSummoner = data.KnownSummoner;
				
				dates = [];
				// Use a hash table as a set to get the full list of metrics.
				metrics = {}
				// Get a list of all of the known dates and metrics. Each metric
				// should become a graph.
				for (var date in data.Records.Daily) {
					snapshot = data.Records.Daily[date]
					dates.push( date );
					
					for (var i = 0; i < snapshot.Stats.length; i++) {
						metric = snapshot.Stats[i]
						metrics[metric.Name] = true;
					}
				}
				
				// Build timeseries data structures
				$scope.timeseries = {}
				for (var metric in metrics) {
					$scope.timeseries[metric] = {}
					
					for (var i = 0; i < dates.length; i++) {
						for (var j = 0; j < data.Records.Daily[dates[i]].Stats.length; j++) {
							m = data.Records.Daily[dates[i]].Stats[j];
														
							if (m.Name == metric) {
								$scope.timeseries[metric][dates[i]] = m.Absolute;								
							}
						}
					}
				}
				
				$scope.metrics = [];
				for (var metric in metrics) {
					$scope.metrics.push(metric);
				}
				
				$scope.dates = dates;
				graphData.setLabels(dates);
				
				for (var i = 0; i < $scope.metrics.length; i++) {
					var graph_data = [];
					
					for (var j = 0; j < dates.length; j++) {
						graph_data.push( $scope.timeseries[$scope.metrics[i]][dates[j]] );
					}
					graphData.setData($scope.metrics[i], graph_data);
				}
				
				graphData.setActiveMetric($scope.metrics[0]);
			});
		}
	});
	
	
	
	/**
	 * Initial configuration and updates to the chart on the page.
	 */
	app.controller("ChartController", function($scope, $http, $rootScope, graphData) {
		$scope.chart = {
			labels : graphData.labels,
			datasets : [{
				fillColor : "rgba(151,187,205,0)",
				strokeColor : "#e67e22",
				pointColor : "rgba(151,187,205,0)",
				pointStrokeColor : "#e67e22",
				data : [0, 0, 0]
			}],
		};
		
		$scope.chart_options = {
			scaleStartValue: 0
		};
		
		$scope.$on("updateGraph", function() {
			console.log("Updating graph...")
			$scope.chart.labels = graphData.getLabels();
			$scope.chart.datasets[0].data = graphData.getData();
		});
	});
})();
