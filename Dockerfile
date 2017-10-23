FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN go-wrapper download
RUN go-wrapper install

CMD ["/app/redis-proxy"]
