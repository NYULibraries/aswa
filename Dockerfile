FROM golang:1.21.2 as builder

RUN update-ca-certificates

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .

FROM alpine:latest

RUN addgroup --gid 1000 docker && \
    adduser --uid 1000 --ingroup docker --disabled-password --gecos "" docker
# Update package list and install curl and jq
RUN apk update && apk add curl jq

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/config /config
COPY --from=builder /app/app /aswa
COPY --from=builder /app/entrypoint.sh /entrypoint.sh

USER docker
ENTRYPOINT ["/entrypoint.sh"]
CMD [ "/aswa" ]
