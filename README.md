A simple redis backed url shortener

Have a redis server either on localhost:6379 or put an address into the environment variable `REDIS_ADDRESS`


```
go run main.go
```

Then you can

```
22:11 $ curl -i -X POST -H 'Content-Type: application/json' localhost:8080/manage --data-raw '{"long_link":"https://cnn.com"}'

HTTP/1.1 200 OK
Date: Thu, 27 Mar 2025 03:20:38 GMT
Content-Length: 28
Content-Type: text/plain; charset=utf-8

{"short_link":"WIjyMkyupl"}
```

Then go to localhost:8080/<<whatever you got back from `short_link`>>

