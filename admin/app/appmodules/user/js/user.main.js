angular.module('user',[
    'profile',
    'user.controllers'
])
    .config(['$stateProvider', function ($stateProvider) {
        $stateProvider
            .state('user', {
                url: '/user',
                templateUrl: '/appmodules/user/views/index.html',
                data: { title: "User Index"}
            })
            .state('user.all', {
                url: '/all',
                templateUrl: '/appmodules/user/views/all.html',
                data: { title: 'All Users' },
                controller: 'UserAll'
            })
    }]);