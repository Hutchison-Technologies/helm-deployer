FROM golang:1.12.4-alpine3.9 as base

ENV GO111MODULE=on
RUN apk add --update git
WORKDIR /usr/src

COPY . ./

FROM base as test
ENTRYPOINT [ "go" ]
CMD [ "test" ]

FROM base
RUN go build
ENTRYPOINT [ "./main" ]