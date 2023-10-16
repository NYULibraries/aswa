FROM golang:1.21.3 as builder

RUN update-ca-certificates

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .

FROM alpine:latest

RUN addgroup -g 1000 docker && \
    adduser -D -u 1000 -G docker docker

# Update package list and install curl and jq
RUN apk update && apk add --no-cache curl jq

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/config /config
COPY --from=builder /app/app /aswa
COPY --from=builder /app/entrypoint.sh /entrypoint.sh

RUN chown -R docker:docker /config && \
    chown docker:docker /aswa && \
    chown docker:docker /entrypoint.sh

USER docker
ENTRYPOINT ["/entrypoint.sh"]
CMD [ "/aswa" ]
