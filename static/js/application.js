var tvApp = angular.module('tvApp', []);

tvApp.controller('TvCtrl', function ($scope, $http) {
    $scope.currentChannel = 0;

    $scope.switchChannel = function (no) {
        $http.get("/channel?no=" + no.toString());
        $scope.query = "";
        $scope.currentChannel = no;
    }

    $http.get("/channels").success(function (data) {
        $scope.channels = data.channels;
        $scope.currentChannel = data.current;
    });
});
