# CORS middleware

Example to run Atreugo server with CORS middleware.

### Usage
CORS is disabled by default.

You could use *NewCorsMiddleware* in *server.UseBefore*

```
...
server := atreugo.New(config)

server.UseBefore(func(ctx *atreugo.RequestCtx) error {
    widthCors := middlewares.NewCorsMiddleware(middlewares.CorsOptions{
        // if you leave allowedOrigins empty then atreugo will treat it as "*"
        AllowedOrigins: []string{"http://localhost:63342", "192.168.3.1:8000", "APP"},
        // if you leave allowedHeaders empty then atreugo will accept any non-simple headers
        AllowedHeaders: []string{"Content-Type", "content-type"},
        // if you leave this empty, only simple method will be accepted
        AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
        AllowCredentials: true,
        AllowMaxAge:      5600,
    })
    widthCors.CorsMiddleware(ctx)

    return ctx.Next()
})
...
```

Or you could use DefaultCors
```
// Use CORS with out options
server.UseBefore(func(ctx *atreugo.RequestCtx) error {
    withCors := middlewares.DefaultCors()
    withCors.CorsMiddleware(ctx)
    
    return ctx.Next()
})
```

### Quick Test

1. Run Atreugo server:
`
go build main.go
`
2. Try *quick_test.html*

### Reference
[fasthttpcors](https://github.com/AdhityaRamadhanus/fasthttpcors)