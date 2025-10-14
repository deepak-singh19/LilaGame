FROM heroiclabs/nakama-pluginbuilder:3.26.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build --trimpath --buildmode=plugin -o ./backend.so

FROM heroiclabs/nakama:3.26.0

COPY --from=builder /backend/backend.so /nakama/data/modules

# Expose ports
EXPOSE 7349 7350 7351

# Create a startup script
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'echo "Starting Nakama with DATABASE_URL: $DATABASE_URL"' >> /start.sh && \
    echo 'echo "Waiting for database to be ready..."' >> /start.sh && \
    echo 'sleep 5' >> /start.sh && \
    echo '/nakama/nakama migrate up --database.address "$DATABASE_URL"' >> /start.sh && \
    echo 'exec /nakama/nakama --database.address "$DATABASE_URL"' >> /start.sh && \
    chmod +x /start.sh

CMD ["/start.sh"]
