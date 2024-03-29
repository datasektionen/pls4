# Running
## Install dependencies
Download a go compiler, at least version 1.22

Download the correct version of [templ](https://templ.guide/) using:
```sh
go install github.com/a-h/templ/cmd/templ@$(grep -oPm1 'github.com/a-h/templ \K[^ ]*' go.sum)
```

## Setup a database
Start a postgresql database, using e.g.:
```sh
docker run -d --name pls4-db -p 5432:5432 -e POSTGRES_PASSWORD=pls4 -e POSTGRES_DB=pls4 -e POSTGRES_USER=pls4 postgres:16-alpine3.19
```
...or add a user and database to an existing instance:
```sql
CREATE USER pls4 WITH PASSWORD 'pls4';
CREATE DATABASE pls4 WITH OWNER pls4;
```

## Setup the login system
Either:
- set up [login](https://github.com/datasektionen/login) locally (not easy),
- get a login api key (probably not easy),
- or set up [nyckeln under dörrmattan](https://github.com/datasektionen/nyckeln-under-dorrmattan) (easy).

## Set up environment variables.
See `.env.example`.

I recommend installing [direnv](https://direnv.net/) and running
```sh
cp .env.example .env
echo "dotenv" > .envrc
echo ".envrc" >> .git/info/exclude
```

## Build/Run
To run and rebuild when code has changed, download [air](https://github.com/cosmtrek/air) and run:
```sh
air -build.pre_cmd="go generate ./..." -build.exclude_regex=".*_templ.go" -build.include_ext="go,templ"
```

To build the project as a binary, run:
```sh
go generate ./...
go build .
```

You can also build an OCI image using the `Dockerfile`.

## Profit
Open https://localhost:3000/ in your web browser!
