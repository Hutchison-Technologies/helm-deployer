# Helm-deployer

## Development

For guidance on how to develop this project, refer to the appropriate subsection below.

### Requirements

First, install [golang](https://golang.org/dl/). Make sure you read about, and correctly set up, your [go workspace](https://golang.org/doc/code.html#Workspaces).

Second, install [dep](https://golang.github.io/dep/), a dependency management program for golang.

If you want to make use of the docker image, install [Docker](https://docs.docker.com/install/)

### Setup

Create the correct package directory in your [go workspace](https://golang.org/doc/code.html#Workspaces):

    $ mkdir $GOPATH/src/github.com/Hutchison-Technologies

Checkout the repository into the above directory:

    $ cd $GOPATH/src/github.com/Hutchison-Technologies && git clone git@github.com:Hutchison-Technologies/helm-deployer.git

Enter the root of the project and install the dependencies using [dep](https://golang.github.io/dep/):

    $ cd helm-deployer && dep ensure

### Building

To build the binary, run:

    $ cd $GOPATH/src/github.com/Hutchison-Technologies/helm-deployer && go build

### Installing

To install the binary (meaning you can run it from your terminal anywhere), run:

    $ cd $GOPATH/src/github.com/Hutchison-Technologies/helm-deployer && go install

You should now be able to run the program from anywhere:

    $ helm-deployer

### Running

If you have built the project, simply execute the binary:

    $ $GOPATH/src/github.com/Hutchison-Technologies/helm-deployer/helm-deployer

OR if you have installed the binary:

    $ helm-deployer

### Testing

To run the unit tests, run:

    $ cd $GOPATH/src/github.com/Hutchison-Technologies/helm-deployer && go test ./...

### Help

To print help (after installing), run:

    $ helm-deployer -h
