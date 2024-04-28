FROM alpine:latest

RUN apk --no-cache add git yq

COPY ./install-kubectl.sh .

RUN ./install-kubectl.sh