FROM golang:1.20
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o carbonifer .
CMD ["./carbonifer"]