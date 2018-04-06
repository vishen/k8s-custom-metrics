# Build the application
FROM golang:1.9 as builder

WORKDIR /go/src/github.com/vishen/k8s-custom-metrics
COPY vendor/ vendor
COPY main.go .

#RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Run the application
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /go/src/github.com/vishen/k8s-custom-metrics/main .

ENTRYPOINT ["./main"]
