FROM golang:1.16.2-alpine as builder

RUN apk --update upgrade \
    && apk --no-cache --no-progress add make \
    && rm -rf /var/cache/apk/*

WORKDIR /go/src/ikit-cache

COPY go.mod go.sum ./
RUN GO111MODULE=on go mod download

COPY . .

RUN make consumer

FROM alpine:3.13

COPY --from=builder /go/src/ikit-cache/consumer /usr/local/bin

ENTRYPOINT [ "consumer" ]
