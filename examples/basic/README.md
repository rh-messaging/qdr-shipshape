# Basic example - using Java clients

## Description

The basic example provides a test suite that deploys the qdr-operator, creates
an Interconnect instance (minimal with defaults) and then runs Java sender and
receiver applications to exchange messages through it.

In this example, we have a few components involved.

1. Client applications
1. Client container image
1. Go wrapper for the client image
1. Example test suite

## Client application

It is located in the *client* directory, and it contains simple Java clients (one Sender and one Receiver).

### Client interface

The parameters expected by the client applications are defined, by default, at a property file located
at *src/main/resources/config.properties*. It contains a few variables and those variables can be overridden
by exporting *Environment Variables* using the same names.

This way if no environment variable is customized, the configuration will be fetch from the
properties file, otherwise the environment variable takes higher precedence.

And this class provided a one line json message (last line logged to stdout) that contains an overview of 
client status and it provides the following output:

    {"delivered":1000,"accepted":1000,"released":0,"rejected":0,"errormsg":""} 

The go wrapper amqp client implementation of the *Result* method parses the JSON above into
a golang type that is used to validate the client results.

### Building the client application

To build the client application, you just need to run:

    mvn clean package

or
  
    make build
    
It will produce a *fat jar* that will be added (later) to a container image, so the two client classes
can be executed inside a container (or a Pod on Kubernetes). 

## Container image

It is build based on the OpenJDK 8 image and contains the *fat jar* that was produced by packaging
the java client application. The image is available under the dockerhub org qdrshipshape.

### Building the image

You can simply run:

    make image
    
To publish it you have to be a member of the qdrshipshape organization on docker hub, and to do it
all you need to run is:

    docker tag qdrshipshape/examples-java-basic docker.io/qdrshipshape/examples-java-basic
    docker push docker.io/qdrshipshape/examples-java-basic
 
## Go wrapper for the client

A go wrapper has been implemented to help executing the client image in a Pod, so it has a Builder to produce
the Pod along with its container and environment variables (to customize interface for
invoking the client application).

Its code can be located at *test/javaclient/* directory.

The builder produces an *AMQP client* type that can be deployed (using the *Deploy* method).
It also provides an implementation for the *Result* method that parses the JSON message.

## The example test suite

The example test suite is located at *test/* directory and it contains three files:

1. basic_suite_test.go
   
   Initializes the suite and parses general shipshape command line arguments, which includes reading
   and parsing of the KUBECONFIG environment variable, used to determine the context to be used. It
   also runs all specs from the *test* package.
1. setup.go

    Contains the common blocks to perform setup and teardown of the components used
    for this sample test. It is separate from the test itself to improve readability.
    
    Yet it is very simple as it just creates an instance of the *Framework* and deploys
    a minimum version of an Interconnect.
1. basic_test.go
    
    Contains the logic for the test implementation itself. It deploys a sender and a receiver,
    waits on both to finish and validate the results.

## Running the basic example

First you must have a KUBECONFIG environment variable set and must be logged in to a Kubernetes or OpenShift cluster using a cluster-admin user.
*NOTE:* qdr-shipshape expects you to use go 1.13+.

Next all you have to do is just run the following command:

    ginkgo -v -r examples/basic/test/
    
alternatively you can use *go test* itself, as:

    go test -v ./examples/basic/test/
