FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache build-base sqlite-dev
COPY . .
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64
RUN go build -o user-server ./cmd/server
RUN go build -o invite ./cmd/invite

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates sqlite-libs
COPY --from=builder /app/user-server /app/
COPY --from=builder /app/invite /app/
CMD ["./user-server"]