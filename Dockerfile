FROM golang:alpine AS build
ADD . /go/src/Bocchi-Go/
ARG GOARCH=amd64
ENV GOARCH ${GOARCH}
ENV CGO_ENABLED 0
WORKDIR /go/src/Bocchi-Go
RUN go build .

FROM alpine
COPY --from=build /go/src/Bocchi-Go/Bocchi-Go /bin/Bocchi-Go
WORKDIR /data
CMD Bocchi-Go