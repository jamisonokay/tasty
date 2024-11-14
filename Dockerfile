FROM golang:1.23.3 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

FROM alpine:latest

RUN apk --no-cache add chromium

WORKDIR /root/

COPY --from=builder /app/main .
COPY .env .

EXPOSE 3005

CMD ["./main"]
