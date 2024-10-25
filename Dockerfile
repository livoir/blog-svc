FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -buildvcs=false -o app ./cmd/main.go

FROM alpine:3.20

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/app .

RUN apk --no-cache add ca-certificates tzdata

RUN chown -R appuser:appuser /app && chmod +x app

USER appuser

EXPOSE 8080

CMD ["./app"]
