FROM golang:1.16.6-alpine3.14 as base

RUN apk add --update git curl gcc build-base bash
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR ${GOPATH}/src/github.com/Hutchison-Technologies/helm-deployer

ENV GO111MODULE=on
COPY . ./
RUN go mod download

FROM base as test
ENTRYPOINT [ "go" ]
CMD [ "test" ]

FROM base
RUN go install
ENTRYPOINT [ "helm-deployer" ]