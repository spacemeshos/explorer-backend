FROM golang:1.24.0-alpine AS build
WORKDIR /src
COPY . .
RUN apk add --no-cache gcc musl-dev
RUN go build -o explorer-stats-api ./cmd/api/

FROM alpine:3.21
COPY --from=build /src/explorer-stats-api /bin/
EXPOSE 5000
EXPOSE 5050
EXPOSE 5070
ENTRYPOINT ["/bin/explorer-stats-api"]
