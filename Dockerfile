FROM golang:1.12 as build

ARG SRC_DIR=/go/src/coding_challenge_01

COPY . ${SRC_DIR}
WORKDIR ${SRC_DIR}
ENV GO111MODULE=on
ENV CGO_ENABLED=0
RUN go get .

FROM scratch
COPY --from=build /go/bin/coding_challenge_01 /bin/
ENTRYPOINT [ "/bin/coding_challenge_01" ]
EXPOSE 80