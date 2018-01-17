function placeController($scope, data, $timeout, $http, $stateParams, $location, $anchorScroll, $uibModal, utils) {
    $scope.$stateParams = $stateParams
    $scope.showSearchBox = true
    $scope.selected = null
    $scope.connection = {}
    $scope.shaked = null
    $scope.$on('$destroy', function () {
        $scope.isDestroed = true
    });

    $scope.merge = function (oldData, newData) {
        var result = {}
        result.name = newData.name
        result.lastUpdated = newData.lastUpdated
        result.serviceMap = {}

        //merging
        for (var serviceName in newData.serviceMap) {
            if (newData.serviceMap.hasOwnProperty(serviceName)) {
                if (!result.serviceMap[serviceName]) {
                    result.serviceMap[serviceName] = {}
                }
                if (oldData.serviceMap && oldData.serviceMap.hasOwnProperty(serviceName)) {
                    result.serviceMap[serviceName].error = newData.serviceMap[serviceName].error
                    result.serviceMap[serviceName].lastUpdated = newData.serviceMap[serviceName].lastUpdated
                    // merge
                } else {
                    result.serviceMap[serviceName] = newData.serviceMap[serviceName]
                }
            }
        }

        // flatting services
        result.services = []
        for (var serviceName in result.serviceMap) {
            if (result.serviceMap.hasOwnProperty(serviceName)) {
                result.services.push(result.serviceMap[serviceName])
            }
        }
        // flatting routes
        for (var i = 0; i < result.services.length; i++) {
            var service = result.services[i]
            service.routes = []
            for (var routeEntry in service.routeMap) {
                if (service.routeMap.hasOwnProperty(routeEntry)) {
                    service.routes.push(service.routeMap[routeEntry])
                }
            }
        }
        $scope.connection.lastUpdated = result.lastUpdated
        return result
    }

    $scope.data = $scope.merge({}, data)
    utils.processSchema($scope.data)
    $scope.graph = utils.buildGraphFromSchema($scope.data)

    $scope.listen = function (delay) {
        $timeout(function () {
            $http({
                method: 'GET',
                url: '/data?env=' + $stateParams.component + "&t=" + new Date().getTime(),
                headers: {'Content-Type': 'application/json;charset=UTF-8'}
            })
                .success(function (newData) {
                    console.info("Has come new data: " + newData)
                    $scope.connection.error = null
                    var data = $scope.merge({}, newData)
                    utils.processSchema(data)
                    utils.updateGraph($scope.graph, data)
                })
                .error(function (error, error2, error3) {
                    $scope.connection.error = "Connection with server is lost"
                });
            if (!!$scope.isDestroed) return
            $scope.listen(10000)
        }, delay, true);
    }
    $scope.listen(5000)

    $scope.drawGraph = function (nodes, edges) {
        console.info("Start draw")
        var data = {
            nodes: nodes,
            edges: edges
        };
        var options = {
            nodes: {
                shape: 'dot',
                size: 10,
                shadow: true
            },
            edges: {
                shadow: true,
                smooth: true,
                arrows: {to: true},
                scaling: {
                    customScalingFunction: function (min,max,total,value) {
                        return value/total;
                    },
                    min:1,
                    max:30
                },
                arrowStrikethrough: false
            }
        };
        var container = document.getElementById('camel-map');
        var graphNetwork = new vis.Network(container, data, options);
        graphNetwork.on("click", function (params) {
            if (params.edges.length > 0 || params.nodes.length > 0 ) {
                var data = {
                    nodes: nodes.get(params.nodes),
                    edges: edges.get(params.edges)
                }
                //prepare
                if (data.nodes.length > 0) {
                    data.endpoint = data.nodes[0].label
                } else {
                    data.route = data.edges[0].route
                }
                $scope.$apply(function () {
                    $scope.selectGraphElement(data)
                });
            }
        });

        if (false) {
            var updateFrameVar = setInterval(function() { updateFrameTimer(); }, 60);
            function updateFrameTimer() {
                graphNetwork.redraw();
                currentRadius += 0.05;
            }
            var currentRadius = 0;
            graphNetwork.on("beforeDrawing", function(ctx) {
                var inode;
                var nodePosition = graphNetwork.getPositions();
                var arrayLength = nodes.length;
                for (inode = 0; inode < arrayLength; inode++) {
                    var node = nodes.get(inode)
                    if (node && node.strike) {
                        ctx.strokeStyle = node.color;
                        ctx.fillStyle = '#fffbfb';

                        var radius = Math.abs(50 * Math.sin(currentRadius + inode / 50.0));
                        ctx.circle(nodePosition[node.id].x, nodePosition[node.id].y, radius);
                        ctx.fill();
                        ctx.stroke();
                    }
                }
            });
        }
        console.info("Finish draw")
        return graphNetwork
    }
    $scope.graphNetwork = $scope.drawGraph($scope.graph.nodesDataSet, $scope.graph.edgesDataSet)

    // -----
    // ----- VIEW
    // -----
    $scope.goToPlaces = function () {
        $location.path('#')
    }

    //Select who must be selected
    if ($location.search().select) {
        if (!$scope.data) return
        if (!$scope.data.environments) return


        for (var i = 0; i < $scope.info.app.components.length; i++) {
            var url = window.decodeURIComponent($location.search().select)
            if ($scope.info.app.components[i].url == url) {
                // $scope.selected = $scope.babies[i]
                $anchorScroll();
                $scope.shaked = $scope.info.app.components[i]
                $location.hash('component_' + url);
                break
            }
        }
    }

    $scope.selectedTab = 'graph'
    $scope.setTab = function (v) {
        $scope.selectedTab = v
    }
    $scope.showHideSearchBox = function() {
        $scope.showSearchBox = !$scope.showSearchBox
        if ($scope.showSearchBox) {
            $scope.setTab('graph')
        }
    }

    //Selected graph endpoint
    $scope.endpointSearchValue = ''
    $scope.selectedGraphEndpoint = null
    $scope.selectGraphEndpoint = function (entry) {
        $scope.selectedGraphEndpoint = $scope.selectedGraphEndpoint == entry ? null : entry
    }
    $scope.$watch('selectedGraphEndpoint', function (newValue, oldValue) {
        var options = {
            // scale: scale,
            // offset: {x:offsetx,y:offsety},
            animation: {
                duration: 1000,
                easingFunction: 'easeOutQuart'
            }
        };

        if (newValue) {
            var id = $scope.graph.endpoints[newValue]
            $scope.graph.nodesDataSet.update([{id: id, size: 20}]);
            $scope.graphNetwork.focus(id, options);
        } else {
            // fly away
            // $scope.graphNetwork.fit({animation: options});
        }
        if (oldValue) {
            var id = $scope.graph.endpoints[oldValue]
            $scope.graph.nodesDataSet.update([{id: id, size: 10}]);
        }
    });

    //Selected graph element
    $scope.selectedGraphElement = null
    $scope.selectGraphElement = function (entry) {
        console.info(entry)
        if (entry.endpoint) {
            $scope.selectGraphEndpoint(entry.endpoint)
        } else {
            $scope.selectGraphEndpoint(null)
        }
        $scope.selectedGraphElement = $scope.selectedGraphElement == entry ? null : entry
        $scope.showOptionsDialog(entry)
    }

    $scope.showOptionsDialog = function (data) {
        var options = {}
        var modalInstance = $uibModal.open({
            templateUrl: 'views/options.html',
            controller: optionsDialogController,
            size: "lg",
            resolve: {
                data: function ($q, $http) {
                    var deferred = $q.defer();
                    deferred.resolve(data)
                    return deferred.promise;
                }
            }
        });
        modalInstance.result.then(function (params) {}, function () {});
    }
}