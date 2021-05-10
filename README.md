# How To's

## `app-db`, `app-es`

Main application service, usually contains business needs

## `services`

External services which are needed for current project. Consists of docker-compose scripts

  - Run via `$ ./services/run.sh` for easier execution

### `services/fixtures`

Sample fixtures

  - Run ES index with `$ curl -XPUT "localhost:9200/recipes?pretty" -H 'Content-Type: application/json' --data-binary "@services/fixtures/recipes.json"`

  - Run DB records with `$ psql -f services/fixtures/recipes.sql postgres://postgres@localhost:5432`

Sample ES queries

  - Run query with `$ curl -XGET "localhost:9200/recipes/_search?pretty" -H 'Content-Type: application/json' --data-binary "@services/fixtures/{es-query}"`

## `tools`

Scripts to interact with services

  - Run via `$ go run ./tools/{script}.go {args}` since each tools are separated for modularization
