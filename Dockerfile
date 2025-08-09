FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache git

# Cache deps
COPY go.mod ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/urlshort ./cmd/server

FROM scratch
# data dir for BoltDB
VOLUME ["/data"]
ENV PORT=8080
ENV BASE_URL=http://localhost:8080
ENV DB_PATH=/data/urls.db

EXPOSE 8080
COPY --from=build /bin/urlshort /urlshort

ENTRYPOINT ["/urlshort"]
