angular.module('flow').service('utils', function () {
    var getRandomColor = function () {
        var letters = '0123456789ABCDEF';
        var color = '#';
        for (var i = 0; i < 6; i++) {
            color += letters[Math.floor(Math.random() * 16)];
        }
        return color;
    }
    return {
        getRandomColor: getRandomColor,
        processSchema: function (data) {
            // Set color
            for (var i = 0; i < data.services.length; i++) {
                var service = data.services[i]
                service.color = service.color || this.getRandomColor()
                for (var j = 0; j < service.routes.length; j++) {
                    service.routes[j].color = this.getRandomColor()
                }
            }
        },
        buildGraphFromSchema: function (data) {
            var graph = {
                edgeMap: {},
                endpoints: {},
                endpointsKeys: [],
                nodesDataSet: null,
                edgesDataSet: null,

                edgeExists: function (from, to) {
                    var k = from + "_" + to
                    return !!this.edgeMap[k]
                },
                addEndpoint: function (endpoint, service) {
                    var e = this.endpoints[endpoint]
                    if (typeof e == "undefined") {
                        e = this.nodesDataSet.length
                        this.endpoints[endpoint] = e
                        // console.info(service.name + ": add endpoint " + endpoint)
                        this.addNode(e, endpoint, service)
                    } else {
                        this.updateNode(e, endpoint, service)
                    }
                },
                addNode: function (id, endpoint, service) {
                    this.nodesDataSet.add({
                        id: id,
                        label: endpoint,
                        color: service.color
                    })
                },
                updateNode: function (id, endpoint, service) {
                    // do nothing
                },
                getRouteTitle: function (route) {
                    var title = '<b>' + route.name + '</b>'
                        + '<br/>State: ' + (route.state || 'none')
                        + '<br/>Uptime: ' + (route.uptime || '-')
                        + '<br/>LastUpdated: ' + (route.lastUpdated ? moment(route.lastUpdated).fromNow(): "-")
                        + '<br/> ----'
                        + '<br/>exchangesTotal: ' + (route.exchangesTotal || 0)
                        + '<br/>exchangesCompleted: ' + (route.exchangesCompleted || 0)
                        + '<br/>exchangesFailed: ' + (route.exchangesFailed || 0)
                        + '<br/>exchangesInflight: ' + (route.exchangesInflight || 0)
                        + '<br/>maxProcessingTime: ' + (route.maxProcessingTime || 0)
                        + '<br/>minProcessingTime: ' + (route.minProcessingTime || 0)
                        + '<br/>lastProcessingTime: ' + (route.lastProcessingTime || 0)
                        + '<br/>meanProcessingTime: ' + (route.meanProcessingTime || 0)
                        + '<br/>totalProcessingTime: ' + (route.totalProcessingTime || 0)
                        + '<br/>failuresHandled: ' + (route.failuresHandled || 0)
                        + '<br/>redeliveries: ' + (route.redeliveries || 0)
                        + '<br/>startTimestamp: ' + (route.startTimestamp || '-')
                    return title
                },
                addEdge: function (route, from, to, service) {
                    var id = from + "_" + to
                    // console.info(service.name + ": add edge " + id + ": " + route.name)
                    var routeRepresentation = this.getRouteRepresentation(route, service)
                    this.edgesDataSet.add({
                        id: id,
                        from: from,
                        to: to,
                        route: route,
                        color: routeRepresentation.color,
                        title: this.getRouteTitle(route),
                        value: routeRepresentation.weight,
                        dashes: routeRepresentation.dashes,
                        label: route.exchangesTotal ? '' + route.exchangesTotal : null,
                        font: {align: 'top'}
                    })
                    this.edgeMap[id] = 1
                },
                updateEdge: function (route, from, to, service) {
                    var id = from + "_" + to
                    var existedEdge = this.edgesDataSet.get(id)
                    // console.info(service.name + ": update edge " + id + ": " + route.name)
                    if (typeof existedEdge != "undefined" && typeof route != "undefined") {
                        this.nodesDataSet.update({
                            id: to,
                            strike: false
                        })
                        if (this.isNeedUpdate(existedEdge, route)) {
                            // console.info(route.name + " " + existedEdge.route.exchangesTotal + ' -> ' + route.exchangesTotal)
                            var routeRepresentation = this.getRouteRepresentation(route, service)
                            this.edgesDataSet.update([{
                                id: id,
                                from: from,
                                to: to,
                                route: route,
                                color: routeRepresentation.color,
                                title: this.getRouteTitle(route),
                                value: routeRepresentation.weight,
                                dashes: routeRepresentation.dashes,
                                label: route.exchangesTotal ? '' + route.exchangesTotal : null,
                                font: {align: 'top'}
                            }]);
                            this.nodesDataSet.update({
                                id: to,
                                strike: true
                            })
                        }
                    }
                },
                getRouteRepresentation: function (route, service) {
                    var color = service.color
                    var dashes = false
                    var weight = 1 //route.exchangesTotal,
                    if (route.state == 'None') {
                        color = "#b2b2b2"
                        dashes = true
                    } else if (route.state != 'Started') {
                        color = "#ff251e"
                        dashes = true
                    }
                    return {
                        color: {color: color},
                        dashes: dashes,
                        weight: weight
                    }
                },
                isNeedUpdate: function (edge, route) {
                    if (typeof route.exchangesTotal == "undefined"
                        || typeof edge.route == "undefined"
                        || typeof edge.route.exchangesTotal == "undefined") {
                        return false
                    }
                    return route.exchangesTotal != edge.route.exchangesTotal
                        || route.state != edge.route.state
                        || route.uptime != edge.route.uptime
                },
                build: function (data) {
                    this.nodesDataSet = new vis.DataSet([])
                    this.edgesDataSet = new vis.DataSet([])

                    // Created elements from endpoints
                    for (var i = 0; i < data.services.length; i++) {
                        var service = data.services[i]
                        for (var j = 0; j < service.routes.length; j++) {
                            var route = service.routes[j]
                            if (route.endpoints && route.endpoints.inputs) {
                                for (var n = 0; n < route.endpoints.inputs.length; n++) {
                                    var endpoint = route.endpoints.inputs[n]
                                    this.addEndpoint(endpoint, service)
                                }
                            }
                            if (route.endpoints && route.endpoints.outputs) {
                                for (var n = 0; n < route.endpoints.outputs.length; n++) {
                                    var endpoint = route.endpoints.outputs[n]
                                    //TODO duplicates
                                    if (endpoint.indexOf("{{") !== -1) {
                                        continue
                                    }
                                    this.addEndpoint(endpoint, service)
                                }
                            }
                        }
                    }
                    // Create edges
                    for (var i = 0; i < data.services.length; i++) {
                        var service = data.services[i]
                        for (var j = 0; j < service.routes.length; j++) {
                            var route = service.routes[j]
                            if (route.endpoints && route.endpoints.inputs) {
                                for (var n = 0; n < route.endpoints.inputs.length; n++) {
                                    var input = route.endpoints.inputs[n]
                                    if (route.endpoints.outputs) {
                                        for (var m = 0; m < route.endpoints.outputs.length; m++) {
                                            var output = route.endpoints.outputs[m]
                                            var from = this.endpoints[input]
                                            var to = this.endpoints[output]
                                            // console.info(service.name + ": add " + input + ' (' + from + ')' +
                                            //     " to " + output + ' (' + to + ')' + ' route ' + route.name + ' exists ' + this.edgeExists(from, to))
                                            // console.info("pre done " + (from && to && !this.edgeExists(from, to)))
                                            // console.info("pre done " + (from && to && !this.edgeMap[from + '_' + to]))
                                            if (typeof from != "undefined"
                                                && typeof to != "undefined"
                                                && !this.edgeExists(from, to)) {
                                                // console.info("done")
                                                this.addEdge(route, from, to, service)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                    //TODO
                    for (var key in this.endpoints) {
                        this.endpointsKeys.push(key);
                    }
                }
            }
            graph.build(data)
            return graph
        },
        updateGraph: function (graph, data) {
            // nodes
            for (var i = 0; i < data.services.length; i++) {
                var service = data.services[i]
                for (var j = 0; j < service.routes.length; j++) {
                    var route = service.routes[j]
                    if (route.endpoints && route.endpoints.inputs) {
                        for (var n = 0; n < route.endpoints.inputs.length; n++) {
                            var endpoint = route.endpoints.inputs[n]
                            var id = graph.endpoints[endpoint]
                            if (typeof id == "undefined") {
                                graph.addEndpoint(endpoint, service)
                            }
                        }
                    }
                    if (route.endpoints && route.endpoints.outputs) {
                        for (var n = 0; n < route.endpoints.outputs.length; n++) {
                            var endpoint = route.endpoints.outputs[n]
                            //TODO duplicates
                            if (endpoint.indexOf("{{") !== -1) {
                                continue
                            }
                            var id = graph.endpoints[endpoint]
                            if (typeof id == "undefined") {
                                graph.addEndpoint(endpoint, service)
                            }
                        }
                    }
                }
            }
            // edges
            for (var i = 0; i < data.services.length; i++) {
                var service = data.services[i]
                for (var j = 0; j < service.routes.length; j++) {
                    var route = service.routes[j]
                    if (route.endpoints && route.endpoints.inputs) {
                        for (var n = 0; n < route.endpoints.inputs.length; n++) {
                            var input = route.endpoints.inputs[n]
                            if (route.endpoints.outputs) {
                                for (var m = 0; m < route.endpoints.outputs.length; m++) {
                                    var output = route.endpoints.outputs[m]
                                    var from = graph.endpoints[input]
                                    var to = graph.endpoints[output]
                                    //TODO duplicates
                                    if (typeof from != "undefined"
                                        && typeof to != "undefined") {
                                        if (graph.edgeExists(from, to)) {
                                            graph.updateEdge(route, from, to, service)
                                        } else {
                                            graph.addEdge(route, from, to, service)
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
})