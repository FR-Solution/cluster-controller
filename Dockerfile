FROM golang:1.18.3-alpine3.16 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

ARG VERSION

RUN go install -ldflags "-s \
    -X main.Version=${VERSION}" \
    /app/cmd/cluster-controller

FROM alpine:3.15.0

COPY --from=builder /go/bin/cluster-controller /usr/local/bin/cluster-controller

ENTRYPOINT ["/usr/local/bin/cluster-controller"]