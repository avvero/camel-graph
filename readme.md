## Camel-graph
Camel-graph is the viewer for camel routes through jolokia
## Configuration - services.json
```json
{
  "environments": [
    {
      "name": "test",
      "services": [
        {
          "name": "test",
          "url": "http://",
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
./camel-graph -httpPort=8080
```