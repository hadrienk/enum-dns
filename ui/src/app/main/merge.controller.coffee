angular.module 'ui'
.controller 'MergeController', ($http, $scope, NgTableParams, interval, into) ->

  # http://angular-ui.github.io/bootstrap/versioned-docs/0.13.4/
  $scope.interval = angular.copy(interval)
  $scope.into = angular.copy(into)

  #   [  ]
  # [  ]
  $scope.into.upper = Math.max($scope.interval.upper, $scope.into.upper)
  $scope.into.lower = Math.min($scope.interval.lower, $scope.into.lower)

    # move a row from the source array to the destination array
  $scope.moveRecord = (source, destination, record) ->
    destination.push record
    $scope.removeRecord(source, record)

  $scope.removeRecord = (records, record) ->
    records.splice(records.indexOf(record), 1)

  $scope.save = ->
    $http.put("/api/interval/#{$scope.into.lower}:#{ $scope.into.upper}", $scope.into)
    .success(->
      $scope.$close($scope.into)
    )

  $scope.cancel = ->
    $scope.$dismiss("cancel")

  return
