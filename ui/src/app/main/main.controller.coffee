angular.module 'ui'
.controller 'MainController', ($http, $scope, $modal) ->
  'ngInject'
  vm = this

  searchQuery = {}

  $scope.mergeWith = (interval) ->
    $modal.open({
      templateUrl: 'app/main/interval.html'
      controller: 'IntervalController'
      size: "lg"
      resolve:
        interval: ->
          console.log(interval)
          return interval
    })

    return ""

  $scope.searchResult = [
    {
      upper: "123654"
      lower: "123654"
      records: [
        {
          order: 100
          preference: 10
          flag: "U"
          service: "E2U+sip"
          regex: "!^.*$!sip:customer-service@example.com!"
          replacement: "."
        },
        {
          order: 100
          preference: 10
          flag: "U"
          service: "E2U+sip"
          regex: "!^.*$!sip:customer-service@example.com!"
          replacement: "."
        },
        {
          order: 100
          preference: 10
          flag: "U"
          service: "E2U+sip"
          regex: "!^.*$!sip:customer-service@example.com!"
          replacement: "."
        }
      ]
    },
    {
      upper: "123654"
      lower: "123654"
      records: [
        {
          order: 100
          preference: 10
          flag: "U"
          service: "E2U+sip"
          regex: "!^.*$!sip:customer-service@example.com!"
          replacement: "."
        },
        {
          order: 100
          preference: 10
          flag: "U"
          service: "E2U+sip"
          regex: "!^.*$!sip:customer-service@example.com!"
          replacement: "."
        },
        {
          order: 100
          preference: 10
          flag: "U"
          service: "E2U+sip"
          regex: "!^.*$!sip:customer-service@example.com!"
          replacement: "."
        }
      ]
    }
  ]

  search = (query) ->
    # Generate parameters for the searches.
    splitter = /([\d:]*)\s/

    numbers = ///
          ([1-9][0-9]{0,14}) # Number of prefix
          (?::([1-9][0-9]{0,14})){0,1} # Optional end number or prefix
      ///

  return
