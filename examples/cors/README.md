# CORS Setting

Example to run Atreugo server with CORS middleware.

###Usage
CORS is disabled by default.

You could copy and use *cors_middleware.go*

```
...
server := atreugo.New(config)
server.UseBefore(corsMiddleware)
...
```

Specific domain name
```
corsAllowOrigin = "http://localhost:3000"
```

###Quick Test

1.Run Atreugo server:
```
go build main.go cors_middleware.go
```

2.Try *quick_test.html*
