FROM alpine:3.21 AS builder

RUN apk add --no-cache go

WORKDIR /app

COPY . .

RUN go build -o main ./cmd/app

FROM alpine:3.21

WORKDIR /app

RUN apk add curl
COPY --from=builder /app/main .

ENV GIN_MODE=release
CMD ["./main"]