angular.module('account.services', [])
    .factory('AuthInterceptor', ['$q', '$rootScope','$window', function ($q, $rootScope,$window) {
        return {
            request: function(config){
                config.headers = config.headers || {};
                if ($window.sessionStorage.token) {
                    config.headers.Authorization = $window.sessionStorage.token;
                } else if ($window.localStorage.token) {
                    config.headers.Authorization = $window.localStorage.token;
                }
                return config;
            },
            responseError: function (rejection) {
                if ((rejection.status === 401) || (rejection.status === 403)) {
                    $rootScope.$broadcast('Auth:Required');
                } else if (rejection.status === 419) {
                    $rootScope.$broadcast('Auth:Forbidden');
                }
                return $q.reject(rejection);
            }
        }
    }]);
