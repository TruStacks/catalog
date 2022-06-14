FROM golang as builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN mkdir /components && \
    for i in $(find pkg/components -name "config.yaml" | cut -d "/" -f3); do \
        mkdir /components/"$i"; \
        cp pkg/components/"$i"/config.yaml /components/"$i"/config.yaml; \
    done

RUN CGO_ENABLED=0 go build -o /build/main ./server

FROM alpine
COPY --from=builder /build/main /main
copy --from=builder /components /components
CMD ["/main"]