FROM golang:1.22-alpine as builder
RUN apk add --no-progress --no-cache gcc musl-dev
WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -tags musl -ldflags '-extldflags "-static"' -o /main main.go
#
FROM alpine:3
COPY --from=builder main /account/bin/main
ENTRYPOINT ["/account/bin/main"]