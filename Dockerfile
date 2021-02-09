FROM alpine:latest

RUN apk add --update alpine-sdk ninja meson

WORKDIR /root/
RUN git clone https://github.com/Jackojc/wotpp.git
WORKDIR /root/wotpp
RUN meson build
RUN ninja -C build 

FROM golang:alpine
WORKDIR /go/src/app
RUN apk add --no-cache libstdc++
COPY app/ .
COPY --from=0 /root/wotpp/build/w++ .
ENV PATH="/go/src/app:${PATH}"
RUN go get -d -v ./...
RUN go install -v ./...

cmd ["waasup"]
