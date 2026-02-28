FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /mysql-admin .

FROM alpine:3.23
RUN apk --no-cache add ca-certificates
COPY --from=builder /mysql-admin /mysql-admin
EXPOSE 80
ENTRYPOINT ["/mysql-admin"]
