FROM golang as build

ENV CGO_ENABLED=0

WORKDIR  /yuanlimm

ADD main.go /yuanlimm/

RUN go get github.com/jinzhu/now && \
    go get github.com/gin-gonic/gin && \
	go get github.com/go-redis/redis && \
	go build -o yuanlimm-server

FROM alpine:3.7

ENV REDIS_ADDR=""
ENV REDIS_PW=""
ENV REDIS_DB=""
ENV WISH_URL=""
ENV GIN_MODE="release"

COPY --from=build /yuanlimm/yuanlimm-server /usr/bin/

RUN echo "http://mirrors.aliyun.com/alpine/v3.7/main/" > /etc/apk/repositories && \
    apk update && \
    apk add ca-certificates

WORKDIR /

ENTRYPOINT ["yuanlimm-server"]