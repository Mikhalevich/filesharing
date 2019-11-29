FROM golang:latest as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/filesharing

FROM scratch
COPY --from=builder /go/bin/filesharing /go/bin/filesharing
COPY --from=builder /app/config.json /go/bin/config.json
COPY --from=builder /app/templates/html /go/bin/templates/html
COPY --from=builder /app/res /go/bin/res

EXPOSE 8080

WORKDIR /go/bin
ENTRYPOINT ["./filesharing", "--config=config.json"]
