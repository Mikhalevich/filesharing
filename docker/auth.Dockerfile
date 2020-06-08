FROM golang:latest as builder

WORKDIR /app

COPY filesharing-auth-service/go.mod .
COPY filesharing-auth-service/go.sum .
RUN go mod download

COPY filesharing-auth-service/ .
COPY cert_auth/ ./cert_auth/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/filesharing-auth-service

FROM scratch
COPY --from=builder /go/bin/filesharing-auth-service /app/filesharing-auth-service
COPY --from=builder /app/cert_auth/private_key.pem /app/cert/

WORKDIR /app/
ENTRYPOINT ["./filesharing-auth-service"]
