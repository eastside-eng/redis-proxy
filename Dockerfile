FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN go-wrapper download

RUN go build -o redis-proxy .

CMD ["/app/redis-proxy"]
