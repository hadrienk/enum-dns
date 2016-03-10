angular.module 'ui'
.controller 'IntervalController', ($http, $scope, NgTableParams, interval) ->

  # http://angular-ui.github.io/bootstrap/versioned-docs/0.13.4/
  $scope.interval = angular.copy(interval)

  $scope.save = ->
    $http.put("/api/interval/#{$scope.interval.lower}:#{ $scope.interval.upper}", $scope.interval)
      .success(->
        $scope.$close($scope.interval)
    )

  $scope.removeRecord = (row) ->
    $scope.interval = (x for x in array when x != row)

  $scope.cancel = ->
    $scope.$dismiss("cancel")

  return
