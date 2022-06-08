FROM golang as builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN CGO_ENABLED=0 go build -o /build/main ./...

FROM scratch
COPY --from=builder /build/main /main
CMD ["/main"]