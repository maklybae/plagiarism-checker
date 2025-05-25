FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0


RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

COPY genproto/go genproto/go
WORKDIR /build/grpcgw
ADD grpcgw/go.mod .
ADD grpcgw/go.sum .
RUN go mod download
COPY grpcgw/ .

RUN go build -ldflags="-s -w" -o /app/main cmd/grpcgw/main.go


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/UTC /usr/share/zoneinfo/UTC
ENV TZ UTC

WORKDIR /app
COPY --from=builder /app/main /app/main


CMD ["./main"]
