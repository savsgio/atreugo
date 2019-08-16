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
