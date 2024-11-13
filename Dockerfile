
FROM golang:1.22.2-alpine AS builder

# RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api_gateway ./cmd


FROM alpine:3.18

# RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/api_gateway .


EXPOSE 8080

CMD ["./api_gateway"]
