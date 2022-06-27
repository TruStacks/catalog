FROM golang as builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN mkdir /components && \
    for i in $(find pkg/components -name "config.yaml" | cut -d "/" -f3); do \
        mkdir /components/"$i"; \
        cp pkg/components/"$i"/config.yaml /components/"$i"/config.yaml; \
    done
RUN cp pkg/catalog.yaml /catalog.yaml
RUN CGO_ENABLED=0 go build -o /build/main ./cmd

FROM alpine
COPY --from=builder /build/main /main
COPY --from=builder /components /data/components
COPY --from=builder /catalog.yaml /data/config.yaml
CMD ["/main"]