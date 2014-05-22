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
                templateUrl: '/appmodules/profile/views/me.html',
                data: { title: 'My Profile' },
                controller: 'ProfileMe'
            })
            .state('profile.edit',{
                url: '/edit/{userid:[a-f0-9]{24}}',
                templateUrl: '/appmodules/profile/views/edit.html',
                data: { title: 'User\'s Profile'},
                controller: 'ProfileEdit'
            })
    }]);