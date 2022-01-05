##
## Build
##
FROM golang:1.17 AS build

WORKDIR /app

COPY . .
RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go mod tidy
RUN go build -o ./goblog

##
## Deploy
##
FROM ubuntu:20.04

WORKDIR /app

COPY --from=build /app /app

#ENTRYPOINT ["/app/goblog"]
