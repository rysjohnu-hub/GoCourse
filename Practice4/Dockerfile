FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o myapp ./cmd/api/main.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/myapp .

COPY --from=builder /app/database/migrations ./database/migrations

EXPOSE 8080

ENTRYPOINT ["./myapp"]