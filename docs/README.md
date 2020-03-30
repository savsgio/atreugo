# Documentation

The complete documentation to create an atreugo server and its features are explained in go.dev reference.

[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/savsgio/atreugo/v11)

# Aditional Documentation

## Routing

### Named parameters

```go
myserver.GET("/users/{name}", func(ctx *atreugo.RequestCtx) error {
  name := ctx.UserValue("name")

  return ctx.TextResponse(fmt.Sprintf("Name: %s", name))
})
```

As you can see, `{name}` is a _named parameter_. The values are accessible via `RequestCtx.UserValues`. You can get the value of a parameter by using the `ctx.UserValue("name")`.

Named parameters only match a single path segment:

```
Pattern: /user/{user}

 /user/gordon                     match
 /user/you                        match
 /user/gordon/profile             no match
 /user/                           no match

Pattern with suffix: /user/{user}_admin

 /user/gordon_admin               match
 /user/you_admin                  match
 /user/you                        no match
 /user/gordon/profile             no match
 /user/gordon_admin/profile       no match
 /user/                           no match
```

#### Optional parameters

If you need define an optional parameters, add `?` at the end of param name. `{name?}`

#### Regex validation

If you need define a validation, you could use a custom regex for the paramater value, add `:<regex>` after the name. For example: `{name:[a-zA-Z]{5}}`.

**_Optional paramters and regex validation are compatibles, only add `?` between the name and the regex. For example: `{name?:[a-zA-Z]{5}}`._**

### Catch-All parameters

The second type are _catch-all_ parameters and have the form `{name:*}`.
Like the name suggests, they match everything.
Therefore they must always be at the **end** of the pattern:

```
Pattern: /data/{name:*}

 /data/                           match
 /data/something                  match
 /data/subdir/something           match
```
