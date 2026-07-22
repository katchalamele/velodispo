FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /velodispo ./cmd/velodispo

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /velodispo /velodispo
EXPOSE 8081
ENTRYPOINT ["/velodispo"]
