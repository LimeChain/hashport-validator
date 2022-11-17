FROM golang:1.17 as build
WORKDIR /tmp/src/hedera-eth-bridge-validator
COPY . .
RUN go build -o main ./cmd

FROM ubuntu:latest
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y ca-certificates
WORKDIR /src/hedera-eth-bridge-validator
COPY --from=build /tmp/src/hedera-eth-bridge-validator .
COPY ./config/node.yml ./config/node.yml
COPY ./config/bridge.yml ./config/bridge.yml
CMD ["./main"]