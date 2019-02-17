FROM  golang:1.11.5-stretch as builder
COPY . /app
WORKDIR /app
RUN make build

FROM alpine:3.9
COPY --from=builder /app/succession /app/succession
CMD /app/succession