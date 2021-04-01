# api
RESTful API for chiefsend written in Go


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
