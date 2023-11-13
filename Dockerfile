# Build stage
FROM golang:1.21-rc-alpine3.17 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .
RUN apk add --no-cache tzdata && go build -o main .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:3.17

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /app/main /app/
COPY --from=build /app/config.yaml /app/

WORKDIR /app

EXPOSE 8010

CMD ["./main"]
