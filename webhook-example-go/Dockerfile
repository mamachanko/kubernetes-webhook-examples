FROM golang:1.17.0 as build
WORKDIR /go/src/github.com/mamachanko/webhook-example
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build ./main.go

FROM alpine:3.14.2
RUN apk --no-cache add ca-certificates
COPY --from=build  /go/src/github.com/mamachanko/webhook-example /
ENTRYPOINT ["/main"]
