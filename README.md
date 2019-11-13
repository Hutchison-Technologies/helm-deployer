# Helm-deployer

This application deals with the deployment of our helm charts to the appropriate cluster. This go application is hosted within a docker image, in our [hutchison-t/docker-images](https://github.com/Hutchison-Technologies/docker-images) repository, under the [helm-deployer](https://github.com/Hutchison-Technologies/docker-images/tree/master/helm-deployer) section.

This is used by almost all of our Jenkins files for deployment of our services via helm.

If you wish to modify this service, update it here, push to the master branch in github when happy. From here, you must rebuild the helm-deployer docker image and push it to docker hub; follow the instructions within the docker image repository.

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
