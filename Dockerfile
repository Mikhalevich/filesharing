FROM golang:1.18 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/filesharing

FROM scratch
COPY --from=builder /go/bin/filesharing /app/filesharing
COPY --from=builder /app/config.json /app/config.json
COPY --from=builder /app/res /app/res
COPY --from=builder /app/templates/html /app/templates/html

EXPOSE 8080

WORKDIR /app/
ENTRYPOINT ["./filesharing"]
