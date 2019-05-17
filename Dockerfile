ARG PACKAGE_NAME=coding_challenge_03
ARG SRC_DIR=/go/src/${PACKAGE_NAME}

FROM golang:1.12 as dockerize
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

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
COPY --from=dockerize /usr/local/bin/dockerize /bin/
CMD [ "/bin/api_server" ]
EXPOSE 80
