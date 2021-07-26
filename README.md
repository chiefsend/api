# ChiefSend API

RESTful API for ChiefSend written in Go

## Documentation

https://app.swaggerhub.com/apis-docs/chiefsend/ChiefSend/1.0

## Application Architecture

- **Redis**: Is used for temporary storage of the background job queue (set it up yourself beforehand)
- **Database**: Stores information about the shares (set it up yourself beforehand)
- **Media Storage**: Is just a folder in a filesystem
- **API Server**: Takes and processes all the HTTP requests
- **Background Job Worker**: Starts with the API Server and handles slow tasks, like scheduled deletion of a share.

## Environment Variables:

Set them up in a `.env` file in the root of this repository

- `PORT`: the port the api listens to (required, example: 6969).
- `DATABASE_DIALECT`: the database dialect (supported: mysql | postgres | sqlite | mssql | clickhouse)
- `DATABASE_URI`: the dsn string with all details for db connection (required)
- `MEDIA_DIR`: the path where the files should be saved (required, absolute path)
- `REDIS_URI`: redis uri (required, example: localhost:6379)
- `REDIS_DB`: number of redis db (required, valid: 0..15)
- `REDIS_PASSWORD`: redis password (optional, omit if none)
- `BACKGROUND_WORKERS`: number of background workers (optional, default: 5)
- `ADMIN_KEY`: the admin key which is passed as a bearer token to authenticate delete and update operations (required)

## Supported Databases:

Note: The Database has to be created beforehand. The Schema can be created automatically by passing `-auto-migrate=true`
flag to the program at start

- MySQL - example dsn: `user:pass@tcp(127.0.0.1:3306)/ChiefSend?charset=utf8mb4&parseTime=True&loc=Local`
- PostgreSQL - example
  dsn: `host=localhost user=user password=password dbname=ChiefSend port=9920 sslmode=disable TimeZone=Asia/Shanghai`
- SQLite - example dsn: `file::memory:?cache=shared`
- SQL Server - example dsn: `sqlserver://user:pass@localhost:9930?database=ChiefSend`
- Clickhouse - example
  dsn: `tcp://localhost:9000?database=ChiefSend&username=user&password=pass&read_timeout=10&write_timeout=20`

Other databases may work if you use the MySQL or PostgreSQL dialect.

## Building

```
go build -o chiefsend-api .
```

## Testing

```
go test ./...
```

## Running

```
./chiefsend-api
```
