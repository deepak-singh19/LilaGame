FROM heroiclabs/nakama-pluginbuilder:3.26.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend

# Copy backend files from the build context (Railway builds from root)
COPY backend/go.mod backend/go.sum* ./
RUN go mod download && go mod tidy

# Copy the rest of the backend source code
COPY backend/ ./

# Build the plugin
RUN go build --trimpath --buildmode=plugin -o ./backend.so

FROM heroiclabs/nakama:3.26.0

COPY --from=builder /backend/backend.so /nakama/data/modules
COPY --from=builder /backend/config.yml /nakama/data/

# Expose ports
EXPOSE 7349 7350 7351

# Override Nakama's default configuration
RUN rm -f /nakama/data/local.yml

# Create a startup script with debugging
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'echo "Starting Nakama with DATABASE_URL: $DATABASE_URL"' >> /start.sh && \
    echo 'echo "Waiting for database to be ready..."' >> /start.sh && \
    echo 'sleep 10' >> /start.sh && \
    echo 'echo "Running database migrations..."' >> /start.sh && \
    echo '/nakama/nakama migrate up --database.address "$DATABASE_URL"' >> /start.sh && \
    echo 'echo "Starting Nakama server..."' >> /start.sh && \
    echo 'exec /nakama/nakama --database.address "$DATABASE_URL"' >> /start.sh && \
    chmod +x /start.sh

ENTRYPOINT ["/start.sh"]