FROM gcr.io/distroless/base-debian11 AS final

# Set the working directory inside the container.
WORKDIR /app

# Copy the pre-built smpp-gateway binary into the container.
# Assumes `docker build` runs from the project root where the binary exists at this path.
# Adjust the source path if necessary.
# Use --chown=nonroot:nonroot for distroless non-root user.
COPY --chown=nonroot:nonroot smpp-gateway /app/smpp-gateway

COPY .env /app/.env 

# Expose the port the SMPP gateway listens on (based on your .env snippet).
# This informs Docker that the container listens on this port.
EXPOSE 2775

# Set the user to 'nonroot'.
USER nonroot

# Define the command to run when the container starts.
# This executes your smpp-gateway binary.
# Environment variables will be injected via Docker Compose.
CMD ["/app/smpp-gateway"]

