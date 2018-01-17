FROM golang:latest
LABEL maintainer avvero

ADD . /app
WORKDIR /app
RUN go get github.com/fatih/pool
RUN go build -o main .
CMD ["/app/main", "-httpPort=8080", "-serviceUpdateIntervalSeconds=5"]
