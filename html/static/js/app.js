
(function() {
	var app = angular.module("lolstatApp", []);

	function formatNumber(x) {
		return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
	}
	
	// Create a factory that can be used to pass champion data around
	// between controllers. This service is maintained by the team
	// controllers and consumed by the stats controller.
	app.factory("teams", function() {
		var teamService = {};

		teamService.allies = [];
		teamService.enemies = [];

		teamService.setAllies = function(list) {
			teamService.allies = list;
		}
		
		teamService.setEnemies = function(list) {
			teamService.enemies = list;
		}
		
		return teamService;
	});

	// Controller for the full app. Keeps track of state that's useful
	// in many places but not directly displayed anywhere.
	app.controller("AppController", function($scope, $http) {		
		$http.get("static/data/metadata.json").success(function(data) {
			// Angular wants milliseconds, JSON is providing seconds.
			$scope.lastUpdated = data.lastUpdated * 1000; 
			$scope.numGames = formatNumber(data.numGames);
			$scope.championList = data.champions;
			
			console.log("Champions loaded.")

			var championSet = new Bloodhound({
				datumTokenizer: Bloodhound.tokenizers.obj.whitespace('name'),
				queryTokenizer: Bloodhound.tokenizers.whitespace,
				local: $scope.championList
			});
		
			championSet.initialize();
		
			// Enable Twitter's typeahead on the champion input fields. 
			$('.championInput').typeahead({
				hint: true,
				highlight: true,
				minLength: 1
			}, {
				displayKey: 'name',
				templates: {
					empty: "<p>No matching summoners.</p>",
					suggestion: Handlebars.compile(["<img class='ac-img' src='{{ img }}' />",
					  "<div class='ac-block'>",
					  "  <p class='ac-title'>{{ name }}</p>",
					  "  <p class='ac-label'>~{{ games }} games</p>",
					  "  <p class='ac-subtitle'>{{ title }}</p>",
					  "</div>",
					].join('\n'))
				},
				source: championSet.ttAdapter()
			});
		});
	});

	app.controller("TeamController", function($scope, teams, $rootScope) {		
		$scope.team = [];
		$scope.side = null; // side = { ally, enemy }
		
		$scope.nextChampion = "";
		
		// This function is run whenever the user performs a keystroke
		// in the nextChampion model / input field.
		$scope.validateChampion = function() {
			// TODO: *definitely* not the right place for this. I'm not
			// sure where it should go though. ng-init doesn't seem to
			// work but it only needs to happen once.
			if ($scope.side == 'ally') {
				teams.setAllies($scope.team);
			}
			else if ($scope.side == 'enemy') {
				teams.setEnemies($scope.team);	
			}
			
			// Validation
			var valid_obj = null;
			var rejection_reason = null;
			
			for (i = 0; i < $scope.championList.length; i++) {
				if ($scope.championList[i].name.toLowerCase() == $scope.nextChampion.toLowerCase()) {
					valid_obj = $scope.championList[i];
				}
			}

			// If valid, add to data structure.
			if (valid_obj != null && $scope.team.length < 5) {
				$scope.team.push(valid_obj);
				$rootScope.$broadcast('teamUpdate');

				$scope.nextChampion = "";

				if ($scope.team.length > 4) {
					$scope.teamFull = true;
				}
			}
		}
	});
	
	app.controller("StatsController", function($scope, $http, teams) {
		// The TeamController will emit a signal whenever the teams are
		// updated. Whenever they're updated we should fire a request to
		// fetch new stats for the newly defined teams.
		$scope.$on('teamUpdate', function(data) {
			var ally_names = teams.allies.map(function(champ) { return champ.shortname; })
			var enemy_names = teams.enemies.map(function(champ) { return champ.shortname; })
			
			// Retrieve some sample data and format it to be easier to read.
			$http.get("team/?allies=" + ally_names + "&enemies=" + enemy_names).success(function(data) {
				// TODO: check .successful status of query and handle failed cases better.
				$scope.stats = data;

				$scope.stats.results.percent = Math.round( ($scope.stats.results.matching / $scope.stats.results.available) * 1000 ) / 10;
				$scope.stats.results.matching = formatNumber($scope.stats.results.matching)
				$scope.stats.results.available = formatNumber($scope.stats.results.available)
				$scope.stats.results.total = formatNumber($scope.stats.results.total)
			});
		});
	});
})();
