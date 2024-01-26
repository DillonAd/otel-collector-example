FROM golang:1.21 AS build

WORKDIR /app

COPY ./service/go.mod ./service/go.sum ./

RUN go mod download

COPY ./service/*.go .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service

FROM debian AS certs

RUN apt-get update && apt-get install -y ca-certificates

FROM scratch AS deploy

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build /out/service .

CMD ["./service"]