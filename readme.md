### Overview
Problem: managing hundreds of complex bus routes manually has become impractical\
Solution: automatic route graph generation\
Status: finished, article is published

----

## Camel-graph
![camel-graph](https://user-images.githubusercontent.com/884337/50089472-812aca00-0238-11e9-8988-eb70b665ef81.png)

Camel-graph is the viewer for routes in servicemix and camel applications. [Article with intention and result explanation (Russian)](https://habr.com/ru/post/435594/)
## Requirenment 
Viewer work over JMS with help of jolokia, so jolokia is required to be in your application or in servicemix.
## Example
![main-sm](https://user-images.githubusercontent.com/884337/50090052-137f9d80-023a-11e9-8bd3-24df76b7e32f.png)
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
