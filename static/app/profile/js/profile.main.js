angular.module('profile',[
    'account',
    'profile.controllers'
])
    .config(['$stateProvider', function ($stateProvider) {
        $stateProvider
            .state('profile', {
                abstract: true,
                url: '/profile',
                template: '<ui-view/>'
            })
            .state('profile.me',{
                url: '/me',
                templateUrl: '/profile/views/me.html',
                data: { title: 'My Profile' },
                controller: 'ProfileMe'
            })
    }]);