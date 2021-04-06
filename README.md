# ChiefSend API
RESTful API for ChiefSend written in Go

## Documentation
https://app.swaggerhub.com/apis-docs/chiefsend/ChiefSend/1.0

## Application Architecture
- **Redis** Is used for temporary storage of the background job queue (set it up yourself beforehand)
- **Database**: Stores information about the shares (set it up yourself beforehand)
- **SendGrid**: Automatically sends eMails with the shares (set it up yourself beforehand)
- **Media Storage**: Is just a folder in a filesystem with the following structure:
- **API Server**: Takes and processes all the HTTP requests
- **Background Job Worker**: Starts with the API Server and handles slow tasks, like sending mail or scheduled deleting of a share.
- **Reverse Proxy (nginx)**: Takes care of exposing the api to the outside world

## Environment Variables:
- `PORT`: the port the api listens to (example: 6969).
- `DATABASE_URI`: the dsn string with all details for db connection
- `MEDIA_DIR`: the path where the files should be saved (absolute path)
- `SENDGRID_API_KEY`: sendgrid api key
- `SENDGRID_SHARE_TEMPLATE`: template id
- `SENDGRID_SENDER_MAIL`: senders mail (has to be verified in sendgrid)
- `SENDGRID_SENDER_NAME`: senders name (example: ChiefSend)
- `REDIS_URI`: redis uri (example: localhost:6379)

## Supported Databases:
Note: Create a database called "ChiefSend" beforehand

- SQLite - example dsn: `file::memory:?cache=shared`
- PostgreSQL - example dsn: `host=localhost user=user password=password dbname=ChiefSend port=9920 sslmode=disable TimeZone=Asia/Shanghai`
- SQL Server - example dsn: `sqlserver://user:password@localhost:1433?database=ChiefSend`

## Building and deploying the API
```
go build -o chiefsend-api .
```

## Configuring a reverse Proxy (nginx)
In order to have HTTPS we can setup a reverse proxy with nginx.
Example configuration might look like this (Replace the items in <> brackets):
```
server {
    listen 80;
    listen [::]:80;
    server_name <DOMAIN>;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name <DOMAIN>;

    ssl_certificate /etc/letsencrypt/live/<DOMAIN>/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/<DOMAIN>/privkey.pem; # managed by Certbot

    if ($http_host !~ "^<DOMAIN>"){
        rewrite ^(.*)$ $scheme://<DOMAIN>/$1 redirect;
    }

    gzip on;
    gzip_types
        text/plain
        text/css
        text/js
        text/xml
        text/javascript
        application/javascript
        application/json
        application/xml
        application/rss+xml
        image/svg+xml;

    location /api/ {
        proxy_pass http://localhost:<PORT>/;
    }
}
```
Don't forget to set up your SSL certificates.
