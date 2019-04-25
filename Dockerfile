FROM golang:1.12.4-alpine3.9 as base

ENV GO111MODULE=off
RUN apk add --update git curl
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR ${GOPATH}/src/github.com/Hutchison-Technologies/bluegreen-deployer

COPY Gopkg.toml Gopkg.lock *.go ./
RUN dep ensure

FROM base as test
ENTRYPOINT [ "go" ]
CMD [ "test" ]

FROM base
RUN go install
ENTRYPOINT [ "bluegreen-deployer" ]