angular.module('account.services', [])
    .service('User', ['$cacheFactory','$http','$rootScope', function ($cacheFactory,$http,$rootScope) {
        var me = this;
        this.$cache = $cacheFactory('userDetails');

        this.setUserDetails = function (userDetails) {
            this.$cache.put('user',userDetails);
        };

        this.clear = function () {
            this.$cache.remove('user');
        };

        this.getUser = function(){
            return this.$cache.get('user');
        };

        this.getUserId = function(){
            return this.$cache.get('user').id;
        };

        this.getUserRole = function(){
            return this.$cache.get('user').role;
        };

        this.updateStatus = function(){
            $http.get('/api/v1/account/status')
                .success(function(data){
                    me.setUserDetails(data.user);
                    $rootScope.$broadcast('Auth:StateChanged');
                })
        }
    }])
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
                if (rejection.status === 401) {
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