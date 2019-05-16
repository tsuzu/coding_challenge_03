ARG PACKAGE_NAME=coding_challenge_01
ARG SRC_DIR=/go/src/${PACKAGE_NAME}

FROM golang:1.12 as build

ARG PACKAGE_NAME
ARG SRC_DIR

COPY . ${SRC_DIR}
WORKDIR ${SRC_DIR}
ENV GO111MODULE=on
ENV CGO_ENABLED=0
RUN go get .

FROM scratch

ARG PACKAGE_NAME

COPY --from=build /go/bin/${PACKAGE_NAME} /bin/api_server
ENTRYPOINT [ "/bin/api_server" ]
EXPOSE 80
