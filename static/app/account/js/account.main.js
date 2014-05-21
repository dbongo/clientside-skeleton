angular.module('account', [
    'ui.router',
    'account.controllers',
    'account.directives',
    'account.services',
    'ui.bootstrap'
])
    .config(['$stateProvider','$httpProvider', function ($stateProvider,$httpProvider) {
        $httpProvider.interceptors.push('AuthInterceptor');

        $stateProvider
            .state('account', {
                url: '/account',
                template: '<ui-view/>'
            })
            .state('account.register', {
                url: '/register',
                templateUrl: '/account/views/register.html',
                data: { title: 'Create Account' },
                controller: 'AccountRegister'
            })
            .state('account.login', {
                url: '/login',
                templateUrl: '/account/views/login.html',
                data: { title: 'Login' },
                controller: 'AccountLogin'
            })
            .state('account.forgot', {
                url: '/forgot',
                templateUrl: '/account/views/forgot.html',
                data: { title: 'Forgot Password' },
                controller: 'AccountForgot'
            })
            .state('account.reset', {
                url: '/reset/{token}',
                templateUrl: '/account/views/reset.html',
                data: { title: 'Verify Password Reset Token' },
                controller: 'AccountReset'
            })
            .state('account.logout', {
                url: '/logout',
                templateUrl: '/account/views/logout.html',
                data: { title: 'Logout' },
                controller: 'AccountLogout'
            });


    }])
    .run(['$rootScope','$state', function($rootScope,$state){
        $rootScope.$on('$stateChangeStart',
            function(event, toState){
                $rootScope.title = toState.data.title;
             });
        $rootScope.$on('Auth:Required', function() {
            $state.go('account.login');
        });
        $rootScope.$on('$locationChangeStart',function(){
            $rootScope.signed = !!window.sessionStorage.token;
        })
    }]);
