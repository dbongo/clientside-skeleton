angular.module('main', [
    'user'
])
    .config(['$locationProvider','$stateProvider','$urlRouterProvider', function ($locationProvider,$stateProvider,$urlRouterProvider) {
        $locationProvider.html5Mode(true);
        $urlRouterProvider.otherwise('/');
        $stateProvider
            .state('home', {
                url: '/',
                templateUrl: '/appmodules/main/views/home.html',
                data: { title: "Home"}
            })
    }]);