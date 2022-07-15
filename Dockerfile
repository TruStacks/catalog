FROM golang as builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN CGO_ENABLED=0 go build -o /build/main ./cmd

FROM alpine
COPY --from=builder /build/main /main
CMD ["/main"]