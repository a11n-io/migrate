# cerberus

For running migrations against [cerberus.a11n.io](cerberus.a11n.io) ([a11n.io](a11n.io))

The cerberus driver will automatically wrap each migration file in an implicit transaction by default. Either all commands in the file succeed, or none succeed.

## Create migrations
Let's create a ResourceType, action and Policy:
Create two files under `cerberusmigrations` folder:
- 000001_create_policy.down.txt
- 000001_create_policy.up.txt

In the `.up.txt` file let's create the artifacts:
```
(crt "Account")
(ca "Account" "Read" "Delete" "View")
(cp "ManageAccount")
```
And in the `.down.txt` let's delete them:
```
(drt "Account")
(dp "ManageAccount")
```

Note that we don't have to delete Actions, since they will be deleted when the Resource Type is deleted.

## Run migrations within your Go app
At the moment, it's not possible to run migrations using the migrate cli.
Here is a very simple app running migrations for the above configuration:
```
import (
    "log"
    
    cerberus "github.com/a11n-io/go-cerberus"
    "github.com/golang-migrate/migrate/v4"
    cerberusmigrate "github.com/golang-migrate/migrate/v4/database/cerberus"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
    cerberusHost := "https://cerberus-api.a11n.io"
    cerberusApiKey := "my api key"
    cerberusApiSecret := "my api secret"
    
    cerberusClient := cerberus.NewClient(cerberusHost, cerberusApiKey, cerberusApiSecret)

    cdriver, err := cerberusmigrate.WithInstance(cerberusClient, &cerberusmigrate.Config{})
    if err != nil {
        log.Fatalf("could not get cerberus driver: %v", err.Error())
    }
    cm, err := migrate.NewWithDatabaseInstance(
        "file://cerberusmigrations", "cerberus", cdriver)
    if err != nil {
        log.Fatalf("could not get cerberus migrate: %v", err.Error())
    } else {
        if err := cm.Up(); err != nil {
            log.Println(err)
        }
        log.Println("cerberus migration done")
    }
}
```