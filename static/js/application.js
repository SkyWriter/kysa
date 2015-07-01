var tvApp = angular.module('tvApp', []);

tvApp.controller('TvCtrl', function ($scope, $http) {
    $scope.currentChannel = -1;

    $scope.switchChannel = function (no) {
        $.post("/channel", { "no": no });
        $scope.query = "";
        $scope.currentChannel = no;
    }

    $scope.turnOff = function() {
        $.post("/off", { "ios_please_dont_cache_me": Math.random() });
        $scope.currentChannel = -1;
    }

    $http.get("/channels").success(function (data) {
        $scope.channels = data.channels;
        $scope.currentChannel = data.current;
    });
});
