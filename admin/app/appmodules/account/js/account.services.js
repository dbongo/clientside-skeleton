angular.module('account.services', [])
    .factory('AuthInterceptor', ['$q', '$rootScope','$window', function ($q, $rootScope,$window) {
        return {
            request: function(config){
                config.headers = config.headers || {};
                if ($window.sessionStorage.token) {
                    config.headers.Authorization = $window.sessionStorage.token;
                }
                return config;
            },
            responseError: function (rejection) {
                if ((rejection.status === 401) || (rejection.status === 403)) {
                    $rootScope.$broadcast('Auth:Required');
                }
                return $q.reject(rejection);
            }
        }
    }]);
/*
    .factory('HttpInterceptor', ['$q', '$rootScope', function ($q, $rootScope) {
        return {
            // On response failure
            responseError: function (rejection) {
                if (canRecover(rejection)) {
                    return responseOrNewPromise
                }
                // Return the promise rejection.
                return $q.reject(rejection);
            }
        };
    }]);
    */