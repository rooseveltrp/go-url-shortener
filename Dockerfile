FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache git

COPY . .

RUN go mod tidy && go mod verify

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /bin/urlshort ./cmd/server

FROM scratch
VOLUME ["/data"]
ENV PORT=8080
ENV BASE_URL=http://localhost:8080
ENV DB_PATH=/data/urls.db
EXPOSE 8080
COPY --from=build /bin/urlshort /urlshort
ENTRYPOINT ["/urlshort"]