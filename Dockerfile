FROM alpine:latest

ARG TARGETPLATFORM

WORKDIR /

COPY build/${TARGETPLATFORM}/certdist /certdist

# Expose the default server port
EXPOSE 8080

RUN chmod +x /certdist

# Run the application once, to ensure we have the correct setup/platform/executable
RUN ./certdist help

# The command to run the application
# The user will need to provide the command-line arguments (e.g., "server", "config.yml")
CMD ["/certdist"]
