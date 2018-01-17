angular.module("flow", [
    'ngRoute',
    'ui.router',
    'ngSanitize',
    'ui.bootstrap',
    'relativeDate'
])
angular.module("flow").constant('constants', {
    version: "1.0.2"
})
// configure our routes
angular.module("flow").config(function ($routeProvider, $stateProvider, $urlRouterProvider) {

    $urlRouterProvider.otherwise("/")

    $stateProvider
        .state('index', {
            url: "/",
            views: {
                "single": {
                    templateUrl: 'views/environments.html',
                    controller: placesController,
                    resolve: placesController.resolve
                }
            }
        })
        .state('component', {
            url: "/listen/:component",
            views: {
                "single": {
                    templateUrl: 'views/environment.html',
                    controller: placeController,
                    resolve: placesController.resolve
                }
            }
        })
})
angular.module("flow").run(function ($rootScope) {

})

angular.module("flow").controller('mainController', function ($scope) {

})

angular.module("flow").filter('toArray', function() { return function(obj) {
    if (!(obj instanceof Object)) return obj;
    var arr = [];
    for (var key in obj) {
        arr.push({ key: key, value: obj[key] });
    }
    return arr;
}});