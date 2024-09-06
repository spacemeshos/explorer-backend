FROM golang:1.23.0-alpine AS build
WORKDIR /src
COPY . .
RUN apk add --no-cache gcc musl-dev
RUN go build -o explorer-stats-api ./cmd/api/
RUN go build -o cache-agent-refresh ./cmd/cache-agent-refresh/

FROM alpine:3.17
COPY --from=build /src/explorer-stats-api /bin/
COPY --from=build /src/cache-agent-refresh /bin/
EXPOSE 5000
EXPOSE 5050
EXPOSE 5070
ENTRYPOINT ["/bin/explorer-stats-api"]
