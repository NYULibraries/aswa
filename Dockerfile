FROM golang:1.18 as builder

RUN update-ca-certificates

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .

FROM alpine:latest

# Update package list and install curl
RUN apk update && apk add curl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/config /config
COPY --from=builder /app/app /aswa
COPY --from=builder /app/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD [ "/aswa" ]
