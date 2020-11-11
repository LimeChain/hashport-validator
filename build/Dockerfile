FROM golang:1.13 as build
WORKDIR /tmp/src/hedera-eth-bridge-validator
COPY . .
RUN go build -o main ./cmd

FROM ubuntu:latest
WORKDIR /src/hedera-eth-bridge-validator
COPY --from=build /tmp/src/hedera-eth-bridge-validator .
COPY ./config/application.yml ./config/application.yml
CMD ["./main"]