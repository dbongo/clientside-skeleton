angular.module('user.controllers',[])
    .controller('UserAll',['$http','$scope', function($http,$scope){
        $scope.users = [];
        $scope.maxSize = 10;
        $scope.currentPage = 1;
        $scope.totalItems = 0;
        $http.get('/api/v1/users')
            .success(function(data){
                $scope.users = data;
                $scope.totalItems = data.length;
            });
    }]);