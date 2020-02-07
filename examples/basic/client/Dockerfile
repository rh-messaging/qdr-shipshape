# Latest openjdk
FROM openjdk:8

LABEL maintainer=fgiorgetti@gmail.com

# Creates a directory to store client app
WORKDIR /opt/client

# Copying scripts
COPY target/*with-dependencies.jar ./

# Runs bash and expect a custom command to be provided at runtime
CMD /bin/bash
