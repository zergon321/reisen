# Use the official Golang image as the base image
FROM golang:1.20.5-alpine3.18 AS builder

# Set the working directory inside the container
WORKDIR /app

# Install necessary packages
RUN apk add ffmpeg

# Copy the Go module files
COPY go.mod go.sum ./

# Download and cache the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vidstreamsplit .