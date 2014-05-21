angular.module('account.controllers', [])
    .controller("AccountRegister", ['$scope', '$http', '$timeout', '$state', function ($scope, $http, $timeout, $state) {
        $scope.user = {
            name: "",
            surname: "",
            email: "",
            email2: ""
        };
        $scope.alerts = [];
        $scope.disabled = false;
        $scope.registerUser = function () {
            $scope.disabled = true;
            $http.put('/api/v1/account', $scope.user)
                .success(function () {
                    $scope.alerts.push({
                        type: "success",
                        msg: "An email to " + $scope.user.email + " has been sent. Please check your email account for login information."
                    });
                    $scope.alerts.push({
                        type: "info",
                        msg: "You will be redirected to the login page in 30 seconds."
                    });
                    $timeout(function () {
                        $state.go('account.login');
                    }, 30000);
                })
                .error(function (data, status) {
                    $timeout(function () {
                        $scope.disabled = false;
                        $scope.alerts = [];
                    }, 3000);
                    if (status == 502) {
                        $scope.alerts.push({
                            type: "danger",
                            msg: "Internal Server Error. Please try again later."
                        });
                    } else {
                        $scope.alerts = data.alerts;
                    }
                });
        };
        $scope.closeAlert = function(index) {
            $scope.alerts.splice(index, 1);
        };
    }])
    .controller("AccountLogin", ['$scope', '$http', '$timeout', '$state','$window', function ($scope, $http, $timeout, $state,$window) {
        $scope.user = {
            email: "",
            password: "",
            rememberme: true
        };
        $scope.alerts = [];
        $scope.disabled = false;

        $scope.loginUser = function () {
            $scope.disabled = true;
            $scope.alerts.push({
                type: "info",
                msg: "Please wait, logging in."
            });
            $http.post('/api/v1/account', $scope.user)
                .success(function (data) {
                    $scope.alerts = [];
                    $window.sessionStorage.token = data.token;
                    $state.go("home");
                })
                .error(function (data, status) {
                    delete $window.sessionStorage.token;
                    $timeout(function () {
                        $scope.disabled = false;
                    }, 3000);
                    if (status == 502) {
                        $scope.alerts.push({
                            type: "danger",
                            msg: "The server is currently unavailable. Please try again later."
                        });
                    } else {
                        $scope.alerts = data.alerts;
                    }
                });
        };
        $scope.closeAlert = function(index) {
            $scope.alerts.splice(index, 1);
        };
    }])
    .controller("AccountLogout", ['$scope', '$http', '$timeout', '$state','$window', function ($scope, $http, $timeout, $state,$window) {
        $scope.alerts = [];
        $http.post('/api/v1/account/logout')
            .success(function (data) {
                delete $window.sessionStorage.token;
                $scope.alerts = data.alerts;
                $timeout(function () {
                    $state.go('home');
                }, 5000);
            })
            .error(function (data, status) {
                $timeout(function () {
                    $scope.disabled = false;
                }, 3000);
                if (status == 502) {
                    $scope.alerts.push({
                        type: "danger",
                        msg: "The server is currently unavailable. Please try again later."
                    });
                } else {
                    $scope.alerts = data.alerts;
                }
            });
        $scope.closeAlert = function(index) {
            $scope.alerts.splice(index, 1);
        };
    }])
    .controller("AccountForgot", ['$scope', '$http', '$timeout', function ($scope, $http, $timeout) {
        $scope.user = {
            email: ""
        };
        $scope.alerts = [];
        $scope.disabled = false;
        $scope.sendRequest = function () {
            $scope.disabled = true;
            $http.post('/api/v1/account/reset', $scope.user)
                .success(function (data) {
                    $scope.alerts = data.alerts;
                })
                .error(function (data, status) {
                    $timeout(function () {
                        $scope.disabled = false;
                    }, 30000);
                    if (status == 502) {
                        $scope.alerts.push({
                            type: "danger",
                            msg: "The server is currently unavailable. Please try again later."
                        });
                    } else {
                        $scope.alerts = data.alerts;
                    }
                });
        };
        $scope.closeAlert = function(index) {
            $scope.alerts.splice(index, 1);
        };
    }])
    .controller("AccountReset", ['$scope', '$http', '$timeout','$state', function ($scope, $http, $timeout, $state) {
        $scope.token = {
            token: $state.params.token
        };
        $scope.alerts = [];
        $scope.disabled = false;
        $scope.resetPassword = function () {
            $scope.disabled = true;
            $http.put('/api/v1/account/reset', $scope.token)
                .success(function (data) {
                    $scope.alerts = data.alerts;
                    $timeout(function () {
                        $state.go('account.login');
                    }, 5000);
                })
                .error(function (data, status) {
                    $timeout(function () {
                        $scope.disabled = false;
                        $scope.alerts = [];
                    }, 10000);
                    if (status == 502) {
                        $scope.alerts.push({
                            type: "danger",
                            msg: "The server is currently unavailable. Please try again later."
                        });
                    } else {
                        $scope.alerts = data.alerts;
                    }
                });
        };
        $scope.closeAlert = function(index) {
            $scope.alerts.splice(index, 1);
        };
    }]);