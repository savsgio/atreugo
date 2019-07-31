Documentation
=============

The complete documentation to create an atreugo server and its features are explained in GoDoc.

See: [GoDoc](https://godoc.org/github.com/savsgio/atreugo)


Aditional Documentation
=======================


Routing
-------


### - URL Parameters

You can add url parameters with `:<param_name>`, for example `/users/:name`.

The parameters could be optional, adding at the end `?` to it `:<param_name>?`, for example `/users/:name/:surname?`

To recover these parameters, just call to `ctx.UserValue("<param_name>")`

```go
myserver.Path("GET", "/users/:name/:surname?", func(ctx *atreugo.RequestCtx) error {
  name := ctx.UserValue("name")
  surname := ctx.UserValue("surname") // Could be nil

  return ctx.TextResponse(fmt.Sprintf("Name: %s, Surname: %s", name, surname))
})
```


### - Fasthttp handler compatibility

Atreugo could execute fasthttp handlers inside views.
To access to `*fasthttp.RequestCtx`, just get it with `ctx.RequestCtx`

```go
func fasthttpHandler(ctx *fasthttp.RequestCtx) {
    ctx.WriteString("Hello world")
}

myserver.Path("GET", "/handler", func (ctx *atreugo.RequestCtx) error {
    fasthttpHandler(ctx.RequestCtx)
    return nil
})
```
