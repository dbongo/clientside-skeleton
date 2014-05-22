angular.module('profile.controllers', [])
    .controller('ProfileMe', ['$rootScope', '$http', '$state', '$scope', '$timeout',
        function ($rootScope, $http, $state, $scope, $timeout) {
            $scope.alerts = [];
            $scope.profile = {};
            $http.get('/api/v1/profile')
                .success(function (data) {
                    $scope.profile = data;
                });

            $scope.updateProfile = function () {
                $http.post('/api/v1/profile', $scope.profile)
                    .success(function (data) {
                        $scope.alerts = data.alerts;
                        $timeout(function () {
                            $scope.alerts = [];
                        }, 5000);
                    })
                    .error(function (data) {
                        $scope.alerts = data.alerts;
                        $timeout(function () {
                            $scope.alerts = [];
                        }, 5000);
                    })
            };
            $scope.closeAlert = function (index) {
                $scope.alerts.splice(index, 1);
            };
        }]);