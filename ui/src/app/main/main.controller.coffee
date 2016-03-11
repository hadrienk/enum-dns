angular.module 'ui'
.controller 'MainController', ($http, $scope, $modal, $q) ->
  'ngInject'
  vm = this

  searchQuery = {}

  $scope.edit = (interval) ->
    $modal.open({
      templateUrl: 'app/main/interval.html'
      controller: 'IntervalController'
      size: "lg"
      resolve: # pass "interval" as ctrl param using the fct.
        interval: ->
          return interval
    })

    return ""

  $scope.split = (interval) ->
    $modal.open({
      templateUrl: 'app/main/split.html'
      controller: 'SplitController'
      size: "lg"
      resolve: # pass "interval" as ctrl param using the fct.
        interval: ->
          return interval
    })

    return ""

  def =
    from: 100000000000000
    to: 999999999999999
    limit: 10

  $http.get("/api/interval", {params: def}).success((data) ->
    $scope.searchResult = data
  ).error((data) ->
    console.log(data)
  )

  canceler = {}

  $scope.search = ->
    if canceler.promise?
      canceler.resolve()
    else
      canceler = $q.defer()

    $http.get("/api/interval", {params: {prefix: $scope.query}, timeout: canceler.promise}).success((data) ->
      $scope.searchResult = data
    ).error((data) ->
      console.log(data)
    ).finally(->
      canceler = {}
    )
    # Generate parameters for the searches.
    splitter = /([\d:]*)\s/

    numbers = ///
          ([1-9][0-9]{0,14}) # Number of prefix
          (?::([1-9][0-9]{0,14})){0,1} # Optional end number or prefix
      ///

  return
