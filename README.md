# How To's

## `app-db`, `app-es`

Main application service, usually contains business needs

## `services`

External services which are needed for current project. Consists of docker-compose scripts

  - Run via `$ ./services/run.sh` for easier execution

### `services/fixtures`

Sample fixtures

  - Run ES index with `$ curl -XPUT "localhost:9200/product?pretty" -H 'Content-Type: application/json' --data-binary "@services/fixtures/product.json"`

  - Run DB records with `$ psql -f services/fixtures/product.sql postgres://postgres@localhost:5432`

## `tools`

Scripts to interact with services

  - Run via `$ go run ./tools/{script}.go {args}` since each tools are separated for modularization
