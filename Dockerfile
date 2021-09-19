FROM golang:1.16.5-alpine3.14 AS builder
ENV GO111MODULE=on
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

# Install OpenSSL
RUN apk add openssl

# Generate TLS Certificates
RUN openssl genrsa -out ./cert/ca.key 4096
RUN openssl req -new -x509 -key ./cert/ca.key -sha256 -subj "/C=US/ST=NY/O=CA, Inc." -days 365 -out ./cert/ca.cert
RUN openssl genrsa -out ./cert/service.key 4096
RUN openssl req -new -key ./cert/service.key -out ./cert/service.csr -config ./cert/certificate.conf
RUN openssl x509 -req -in ./cert/service.csr -CA ./cert/ca.cert -CAkey ./cert/ca.key -CAcreateserial -out ./cert/service.pem -days 365 -sha256 -extfile ./cert/certificate.conf -extensions req_ext

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl -o app ./cmd/grpcserver/main.go

FROM scratch
COPY --from=builder /src/cert /cert
COPY --from=builder /src/app /app
EXPOSE 9000
ENTRYPOINT ["/app"]
