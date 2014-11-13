
/**
 * This function accepts a daily, weekly, or monthly record set and
 * converts it into an array that can be used for generating a timeline.
 * 
 * Example value for 'record' is Records.Daily.
 * Output should look like: [{x: <unix_ts>, y: <value>}, {...}]
 */
function timeline(metric, records) {
	output = [];
	
	for (date in records) {
		console.log(records[date]);
		x = Math.round( new Date(date).getTime() / 1000 ); 
		y = records[date].Stats[metric].Value;
		
		output.push( {x: x, y: y} );
	}
	
	return output;
}

(function() {
    var app = angular.module("LoLStats", []);
    
    /**
	 * The application-level controller for the full app.
	 */
	app.controller("AppController", function($scope, $http) {
		// Identify whether the requestedSummoner is a known entity.
		$scope.validSummoner = false;
		$scope.metrics = [];
		$scope.dates = [];
				
		// This should make a request to get the JSON response for the provided
		// summoner.
		$scope.requestSummoner = function() {			
			$http.get("/summoner/?name=brigado").success(function(data) {		
			//$http.get("summoners/" + $scope.requestedSummoner).success(function(data) {
				$scope.validSummoner = data.KnownSummoner;
				$scope.summonerData = data.Records;
				
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

				console.log("Broadcasting update request");
				$scope.$broadcast("summonerUpdate", null);				
			});
		}
		
		$scope.requestSummoner();
	});
	
	app.controller("SummaryController", function($scope) {
		// Set the default.
		var active_tab = document.getElementById("overview-tab");
		var active_page = document.getElementById("overview-details");
		
		$scope.changeTo = function(target_tab) {
			console.log("Activating " + target_tab);
			// Remove the 'active-pane' class from the currently active page.
			active_tab.className = "summary-pane";
			active_page.className = "detail-view";
			
			// Add the 'active-pane' class to the newly active page.
			active_tab = document.getElementById(target_tab + "-tab");
			active_tab.className += " active-pane";
			active_page = document.getElementById(target_tab + "-details");
			active_page.className += " active-view";			
		}
	});
	
	/**
	 * 
	 */
	app.controller("ReportingController", function($scope, $attrs) {
		console.log("Controller for " + $attrs.metric + " is live");
		$scope.metric = metric_data[$attrs.metric];
		
		// Once we get the summoner's data we hsould update the reporting element.
		$scope.$on("summonerUpdate", function() {
			console.log("Updating " + $attrs.metric);
			// Convert the user's performance data into a time series if this is a
			// chart-based metric.
			tlData = timeline($attrs.metric, $scope.summonerData.Daily);
			$scope.metric.value = tlData[tlData.length - 1].y;
			$scope.metric.context = "above average for your rank";

			// Draw the graph
			var graph = new Rickshaw.Graph( {
				element: document.querySelector("#" + $attrs.metric + "-chart"),
				width: 1450,
				height: 200,
				series: [ {
					color: "steelblue",
					data: tlData,
					name: $scope.metric.name
				} ]
			});
			
			var xaxis = new Rickshaw.Graph.Axis.Time( {graph: graph} );
			var yaxis = new Rickshaw.Graph.Axis.Y( {
				graph: graph,
				orientation: "left",
				tickFormat: Rickshaw.Fixtures.Number.formatKMBT,
				element: document.getElementById($attrs.metric + "-yaxis"),
			});
			var hoverDetail = new Rickshaw.Graph.HoverDetail( {
				graph: graph
			} );

			graph.render();
		});
	});
})();
