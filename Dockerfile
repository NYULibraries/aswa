FROM golang:1.27.3@sha256:f20d83a83e18dbd2a4714b8ffd4b30b287b6243962dd75608e86abcc63ba4efc AS builder

RUN update-ca-certificates

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=1000:1000 /app/config /config
COPY --from=builder --chown=1000:1000 /app/app /aswa

USER 1000:1000
ENTRYPOINT ["/aswa"]
