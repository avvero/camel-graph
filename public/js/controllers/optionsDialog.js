function optionsDialogController(data, $scope, $uibModalInstance) {
    $scope.data = data
    $scope.ok = function () {
        $uibModalInstance.close(data);
    }
}