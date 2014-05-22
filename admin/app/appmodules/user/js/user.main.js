angular.module('user',[
    'profile',
    'user.controllers'
])
    .config(['$stateProvider', function ($stateProvider) {
        $stateProvider
            .state('user', {
                url: '/user',
                templateUrl: '/appmodules/user/views/all.html',
                data: { title: 'All Users' },
                controller: 'UserAll'
            })
    }]);