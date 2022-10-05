# Gets the latest Fedora from dockerhub
FROM fedora:31

LABEL maintainer=fgiorgetti@gmail.com

# Custom variables
ENV CLIENT_CONFIG_DIR /opt/client

RUN dnf -y install \
    ca-certificates \
    python3-qpid-proton \
    wget \
    iputils \
    net-tools \
    && dnf clean all -y

# Creates a directory to store client scripts
WORKDIR /opt/client

# Copying scripts
COPY *.py *.ini ./

# Runs bash and expect a custom command to be provided at runtime
CMD /bin/bash

