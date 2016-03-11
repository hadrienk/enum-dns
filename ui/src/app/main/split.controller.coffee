angular.module 'ui'
.controller 'SplitController', ($http, $scope, $q, interval) ->

  # http://angular-ui.github.io/bootstrap/versioned-docs/0.13.4/
  $scope.interval = angular.copy(interval)

  if $scope.interval.upper - $scope.interval.lower >= 2
    middle = $scope.interval.lower + ($scope.interval.upper - $scope.interval.lower) // 2
    $scope.intervals = [
      {
        upper: middle - 1
        lower: angular.copy($scope.interval.lower)
        records: angular.copy($scope.interval.records)
      }
      {
        upper: angular.copy($scope.interval.upper)
        lower: middle
        records: angular.copy($scope.interval.records)
      }
    ]

  $scope.save = ->
    $q.all(for interval in $scope.intervals
      $http.put("/api/interval/#{interval.lower}:#{interval.upper}", interval)
    ).then(->
      $scope.$close()
    )

  $scope.removeRecord = (row) ->
    $scope.interval = (x for x in array when x != row)

  $scope.cancel = ->
    $scope.$dismiss("cancel")

  return
