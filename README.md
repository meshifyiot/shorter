# shorter

A simple redis backed url shortener. It currently will create short links with 10 url safe characters. The links will work for 1 week, and then they will 404.

## local use

### from command line with an existing redis server

Have a redis server either on localhost:6379 or put an address into the environment variable `REDIS_ADDRESS`

```
go run main.go
```

Then you can

```
12:41 $ curl -i -X POST localhost:8080/manage --data-raw '{"long_link": "https://cnn.com"}'
HTTP/1.1 200 OK
Date: Thu, 07 Aug 2025 17:41:54 GMT
Content-Length: 28
Content-Type: text/plain; charset=utf-8

{"short_link":"xAwU9tpms_"}
```

Then go to localhost:8080/<<whatever you got back from `short_link`>>

```
12:41 $ curl -i localhost:8080/xAwU9tpms_
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: https://cnn.com
Date: Thu, 07 Aug 2025 17:42:07 GMT
Content-Length: 38

<a href="https://cnn.com">Found</a>.
```

### docker compose

A docker compose file is provided that will stand up a valkey server for the backend storage.

```
docker compose build
```

followed by

```
docker compose up
```

Then you can

```
12:41 $ curl -i -X POST localhost:8080/manage --data-raw '{"long_link": "https://cnn.com"}'
HTTP/1.1 200 OK
Date: Thu, 07 Aug 2025 17:41:54 GMT
Content-Length: 28
Content-Type: text/plain; charset=utf-8

{"short_link":"xAwU9tpms_"}
```

Then go to localhost:8080/<<whatever you got back from `short_link`>>

```
12:41 $ curl -i localhost:8080/xAwU9tpms_
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: https://cnn.com
Date: Thu, 07 Aug 2025 17:42:07 GMT
Content-Length: 38

<a href="https://cnn.com">Found</a>.
```
