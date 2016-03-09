angular.module 'ui'
  .controller 'IntervalController', ($http, $scope, NgTableParams, interval) ->

    # http://angular-ui.github.io/bootstrap/versioned-docs/0.13.4/

    $scope.interval = interval

    $scope.save = ->
      console.log("save")
      $scope.$close("success")

    $scope.cancel = ->
      $scope.$dismiss("cancel")

    return
