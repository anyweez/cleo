
(function() {
	var app = angular.module("lolstatApp", []);

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
		$http.get("data/championList.json").success(function(data) {
			$scope.championList = data;
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
			// Retrieve some sample data.
			$http.get("data/stats.json").success(function(data) {
				$scope.stats = data;
				$scope.stats.percent = Math.round( ($scope.stats.matching / $scope.stats.available) * 1000 ) / 10;
				$scope.stats.percent += Math.random();
			});
		});
	});
})();

// X. clicking button adds champion to allies.
// X. when num allies == 5, don't show new_champion
// X. work for enemies
// 4. autocomplete
// 5. make web request when champion is added to array
// 6. update stats when response is received 
