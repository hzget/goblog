FROM golang:1.17

WORKDIR /go/src/app
COPY . .

RUN go env -w GOPROXY="https://goproxy.cn,direct"
RUN go mod tidy
RUN go install

#CMD ["goblog"]
