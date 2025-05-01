FROM gcr.io/distroless/base-debian11 AS final

# Set the working directory inside the container.
WORKDIR /app

COPY --chown=nonroot:nonroot manager-api /app/manager-api

COPY .env /app/.env 
# Environment variables will be managed via Docker Compose or runtime injection.

# Expose the port the manager API listens on, as defined previously.
# This informs Docker that the container listens on this port, but it doesn't publish it automatically.
EXPOSE 8001

# Set the user to 'nonroot'. This is the default in distroless base images,
# but explicitly setting it is good practice for clarity and security.
USER nonroot

# Define the command to run when the container starts.
# This executes your manager-api binary.
CMD ["/app/manager-api"]

