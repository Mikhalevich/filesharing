FROM golang:latest as builder

WORKDIR /app

#COPY go.mod .
#COPY go.sum .
#RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/filesharing ./cmd/filesharing/main.go

FROM scratch
COPY --from=builder /go/bin/filesharing /app/filesharing

EXPOSE 8080

WORKDIR /app/
ENTRYPOINT ["./filesharing"]
