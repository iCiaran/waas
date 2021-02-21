FROM alpine:latest

RUN apk add --update alpine-sdk ninja meson

WORKDIR /root/
RUN git clone https://github.com/Jackojc/wotpp.git
WORKDIR /root/wotpp
RUN meson build -Ddisable_run=true
RUN ninja -C build 
RUN git rev-parse --short HEAD > commit_hash_short
RUN git rev-parse HEAD > commit_hash_long

FROM golang:alpine
WORKDIR /go/src/app
RUN apk add --no-cache libstdc++

ARG USER=wotpp
RUN adduser -D $USER
USER $USER

RUN id

COPY --chown=$USER:$USER app/ .
COPY --chown=$USER:$USER --from=0 /root/wotpp/build/w++ .

COPY --chown=$USER:$USER --from=0 /root/wotpp/commit_hash_short .
COPY --chown=$USER:$USER --from=0 /root/wotpp/commit_hash_long .
RUN ./replace_commit_hashes.sh

ENV PATH="/go/src/app:${PATH}"
RUN go get -d -v ./...
RUN go install -v ./...

cmd ["waas"]
