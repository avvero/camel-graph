## Camel-graph
Camel-graph is the viewer for camel routes through jolokia
![main](https://user-images.githubusercontent.com/884337/50089386-49bc1d80-0238-11e9-86e1-e504b9dfa5ce.png)
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
