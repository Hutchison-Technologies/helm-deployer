FROM golang:1.12.4-alpine3.9 as base

ENV GO111MODULE=on
WORKDIR /usr/src

COPY . ./

FROM base as test
ENTRYPOINT [ "go" ]
CMD [ "test" ]

FROM base
RUN go build
ENTRYPOINT [ "go" ]
CMD [ "test" ]