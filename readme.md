## Camel-graph
Camel-graph is the viewer for camel routes through jolokia

![camel-graph](https://user-images.githubusercontent.com/884337/50089472-812aca00-0238-11e9-8988-eb70b665ef81.png)
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
