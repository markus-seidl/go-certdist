FROM alpine:latest

ARG TARGETPLATFORM

WORKDIR /

COPY build/${TARGETPLATFORM}/certdist /certdist

# Expose the default server port
EXPOSE 8080

# The command to run the application
# The user will need to provide the command-line arguments (e.g., "server", "config.yml")
ENTRYPOINT ["/certdist"]
