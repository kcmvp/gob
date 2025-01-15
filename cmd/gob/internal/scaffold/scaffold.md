
## init command
Initialize scaffold for a project, mainly include following:
* [ ] Import dependencies
* [ ] Project layout, runnable template code and configurations etc.

### BuiltIn
```shell
gob init
```
Execute command without any arguments will do the following initializations:

#### Configurations
Create below configurations in the workspace or project root directory. 
workspace root directory take precedence over project root directory.
* [ ] build.yaml
* [ ] .golangci.yaml

#### Test Frameworks
* [ ] [testify](https://github.com/stretchr/testify)
* [ ] [gofakeit](https://github.com/brianvoe/gofakeit)

#### Build Tools(Plugins) & Flow
* [ ] [golangci-lint](https://github.com/golangci/golangci-lint)
* [ ] [pretty test](https://github.com/gotestyourself/gotestsum)

### Support parameters

#### Build(Development) Flow
* [ ] [GitHub Flow]()

#### Database
* [ ] [mysql](https://github.com/go-sql-driver/mysql)
* [ ] [pg](https://github.com/jackc/pgx)
* [ ] [pg](https://github.com/lib/pq)
* [ ] [sqlite](https://github.com/mattn/go-sqlite3)

#### Web frameworks
* [ ] [gin](https://github.com/gin-gonic/gin)
* [ ] [echo](https://github.com/labstack/echo)
* [ ] [chi](https://github.com/go-chi/chi)
* [ ] [fiber](https://github.com/gofiber/fiber)


## dbo command
