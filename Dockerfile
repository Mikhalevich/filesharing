FROM golang:latest

WORKDIR /go/src/github.com/Mikhalevich/filesharing
COPY . .

RUN go get -d -v ./...
RUN go build

EXPOSE 8080

CMD ./filesharing
