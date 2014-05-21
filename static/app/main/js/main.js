angular.module('main', [
    'profile'
])
    .config(['$locationProvider','$stateProvider', function ($locationProvider,$stateProvider) {
        $locationProvider.html5Mode(true);
        $stateProvider
            .state('home', {
                url: '/',
                templateUrl: '/main/views/home.html',
                data: { title: "Home"}
            })
    }]);