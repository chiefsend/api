# ChiefSend API
RESTful API for chiefsend written in Go

## Application Architectur
- **Redis** Is used for temporary storage of the background job queue (set it up yourself beforehand)
- **Database**: Stores information about the shares (set it up yourself beforehand)
- **SendGrid**: Automatically sends eMails with the shares (set it up yourself beforehand)
- **Media Storage**: Is just a folder in a filesystem with the following structure:
    ```
    ./media/
    ├── prod  # shared files in production 
    │   ├── data
    │   └── temp
    └── test  # files from the test environment
        ├── data
        └── temp
    ```
- **API Server**: Takes and processes all the HTTP requests
- **Background Job Worker**: Starts with the API Server

## Environment Variables:
- `PORT`: the port the api listens to (default: 6969)
- `CHUNK_SIZE`: max chunk size the api processes (default: 10 << 20 (10 MB))   
- `DATABASE_URI`: the dsn string with all details for db connection (defaul: in memory redis database)
- `MEDIA_DIR`: the path where the files should be saved (default: ./media)
- `SENDGRID_API_KEY`: sendgrid api key
- `SENDGRID_SHARE_TEMPLATE`: template id
- `SENDGRID_SENDER_MAIL`: senders mail (has to be verified in sendgrid)
- `SENDGRID_SENDER_NAME`: senders name (default: ChiefSend)
- `REDIS_URI`: redis uri (default: localhost:6379)
- `REDIS_PASSWORD`: redis password (default: empty)

## Supported Databases:
Note: Create a database called "ChiefSend" beforehand

- SQLite - example dsn: `file::memory:?cache=shared`
- PostgreSQL - example dsn: `host=localhost user=user password=password dbname=ChiefSend port=9920 sslmode=disable TimeZone=Asia/Shanghai`
- SQL Server - example dsn: `sqlserver://user:password@localhost:1433?database=ChiefSend`

## Building and deploying the API
TODO

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

    client_max_body_size 100M;
    proxy_buffer_size 1024k;
    proxy_buffers 4 1024k;
    proxy_busy_buffers_size 1024k;

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

    location /api/v1/ {
        include proxy_params;
        proxy_pass http://localhost:<PORT>;
    }
}
```

Don't forget to set up SSL certificates:
```
sudo apt install certbot
sudo systemctl stop nginx
sudo certbot certonly -d <DOMAIN>  //---> chose option 1
sudo systemctl start nginx
```
