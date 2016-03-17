angular.module 'ui'
.controller 'MainController', ($http, $scope, $modal, $q) ->
  'ngInject'
  vm = this

  $scope.params =
    from: 100000000000000
    to: 999999999999999
    limit: 10

  # Opens a modal that allows merging interval and into
  $scope.merge = (interval, into) ->
    $modal.open({
      templateUrl: 'app/main/merge.html'
      controller: 'MergeController'
      size: "lg"
      resolve: # pass "interval" as ctrl param using the fct.
        interval: ->
          return interval
        into: ->
          return into
    }).result.then($scope.search, null)

  # Opens a modal for interval editing
  $scope.edit = (interval) ->
    $modal.open({
      templateUrl: 'app/main/interval.html'
      controller: 'IntervalController'
      size: "lg"
      resolve: # pass "interval" as ctrl param using the fct.
        interval: ->
          return interval
    }).result.then($scope.search, null)

  # Opens a modal that splits the interval in two halves
  $scope.split = (interval) ->
    $modal.open({
      templateUrl: 'app/main/split.html'
      controller: 'SplitController'
      size: "lg"
      resolve: # pass "interval" as ctrl param using the fct.
        interval: ->
          return interval
    }).result.then($scope.search, null)

  $scope.previous = ->
    [{lower}, ...] = $scope.searchResult
    $scope.params.before = lower - 1
    $scope.params.after = null
    $scope.search()

  $scope.next = ->
    [..., {upper}] = $scope.searchResult
    $scope.params.after = upper + 1
    $scope.params.before = null
    $scope.search()

  normalizeNumber = (num) ->
    return unless num?
    num = num + ''
    num = num.concat '0' for [1..15 - num.length] if 15 - num.length
    return num * 1

  canceler = {}

  $scope.searching = false

  $scope.search = ->
    if $scope.query == ""
      from = 100000000000000
      to = 999999999999999
    else
      # Compute from and to.
      re = /([1-9][0-9]{0,14})(?:[\s,;:-]?([1-9][0-9]{0,14}))?/
      return unless re.test($scope.query)

      [..., from, to] = re.exec($scope.query)
      to = if to? then normalizeNumber(to) else normalizeNumber(1 * from + 1) - 1
      from = normalizeNumber(from)

    [$scope.params.from, $scope.params.to] = [from, to]

    $scope.searching = true
    if canceler.promise?
      canceler.resolve()
    else
      canceler = $q.defer()

    $http.get("/api/interval", {params: $scope.params, timeout: canceler.promise}).success((data) ->
      $scope.searchResult = data
      [{lower: $scope.min}, ..., {upper: $scope.max}] = data
    ).error(->
      $scope.searchResult = null
    ).finally(->
      canceler = {}
      $scope.searching = false
    )

  return
