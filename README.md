# cos

A cos disk driver for facades.Storage of Goravel.

## Install

1. Add package

```
go get -u github.com/goravel/cos
```

2. Register service provider

```
// config/app.go
import "github.com/goravel/cos"

"providers": []foundation.ServiceProvider{
    ...
    &cos.ServiceProvider{},
}
```

3. Publish configuration file
dd
```
go run . artisan vendor:publish --package=github.com/goravel/cos
```

4. Fill your cos configuration to `config/cos.go` file

5. Add cos disk to `config/filesystems.go` file

```
// config/filesystems.go
import (
    "github.com/goravel/framework/contracts/filesystem"
    cosfacades "github.com/goravel/cos/facades"
)

"disks": map[string]any{
    ...
    "cos": map[string]any{
        "driver": "custom",
        "via": func() (filesystem.Driver, error) {
            return cosfacades.Cos(), nil
        },
    },
}
```

## Testing

Run command below to run test(fill your owner cos configuration):

```
TENCENT_ACCESS_KEY_ID= TENCENT_ACCESS_KEY_SECRET= TENCENT_BUCKET= TENCENT_URL= go test ./...
```
