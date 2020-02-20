# CORS middleware

Example to run Atreugo server with CORS middleware.

### Usage
CORS is disabled by default.

You could use *NewCorsMiddleware* in *server.UseAfter*

```
...
server := atreugo.New(config)

cors := middlewares.NewCorsMiddleware(middlewares.CorsOptions{
    AllowedOrigins:   []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
    AllowedHeaders:   []string{"Content-Type", "content-type"},
    AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
    ExposedHeaders:   []string{"Content-Length, Authorization"},
    AllowedVary:      []string{"Origin, User-Agent"},
    AllowCredentials: true,
    AllowMaxAge:      5600,
})

server.UseAfter(cors)
...
```

### Quick Test

1. Run Atreugo server:
`
go build main.go
`
2. Try *quick_test.html*

### Reference
[fasthttpcors](https://github.com/AdhityaRamadhanus/fasthttpcors)