FROM golang:1.16-alpine

# install better process initializer (dumb-init by yelp)
RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64
RUN chmod +x /usr/bin/dumb-init

# copy binary
COPY chiefsend-api /

# define entrypoint
ENTRYPOINT ["/usr/bin/dumb-init", "/chiefsend-api"]
