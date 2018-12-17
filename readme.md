## Camel-graph
Camel-graph is the viewer for camel routes through jolokia
## Configuration - services.json
```json
{
  "environments": [
    {
      "name": "dev",
      "services": [
        {
          "name": "smx",
          "url": "http://localhost:8181",
          "color": "#62aa34",
          "authorization": {
            "login": "smx",
            "pass": "smx"
          }
        }
      ]
    }
  ]
}

```
## Launch
```
go build
./camel-graph -httpPort=8080
```