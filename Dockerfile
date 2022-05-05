#build stage
FROM golang:1.16.13-alpine3.15 AS builder
# Add a work directory
WORKDIR /app
# Cache and install dependencies
COPY go.mod go.sum ./
RUN go mod download
# Copy app files
COPY . .

# Build go binary file
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .
# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist
COPY ./config/config.yml .
# Copy binary from build to main folder
RUN cp /app/main .


#final stage
FROM alpine:3.15.0
RUN apk --no-cache add ca-certificates
COPY --from=builder /dist/main /main
COPY ./config/config.yml ./config/
COPY ./static ./static/
ENTRYPOINT /main
EXPOSE 3000