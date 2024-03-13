FROM alpine:3.19 as tdlib_builder

RUN apk add --no-cache alpine-sdk linux-headers git zlib-dev openssl-dev gperf php cmake

WORKDIR /tmp/build_tdlib/

RUN git clone https://github.com/tdlib/td.git /tmp/build_tdlib
WORKDIR /tmp/build_tdlib/td
RUN git checkout 1a50ec4
RUN mkdir build
WORKDIR /tmp/build_tdlib/build/

RUN cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local ..
RUN cmake --build . --target install

FROM golang:1.22.1-alpine3.19 as builder
RUN apk add --no-cache git openssl-dev zlib-dev build-base linux-headers
COPY --from=tdlib_builder /usr/local/lib/libtd* /usr/local/lib/
COPY --from=tdlib_builder /usr/local/include/td /usr/local/include/td

WORKDIR /build
WORKDIR /build/simple-telegram-forwarder
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build --trimpath --ldflags "-s -w"

FROM alpine:3.19

RUN apk add --no-cache tzdata ca-certificates libstdc++
WORKDIR /app
COPY --from=builder /build/simple-telegram-forwarder/simple-telegram-forwarder .
CMD ["./simple-telegram-forwarder"]
