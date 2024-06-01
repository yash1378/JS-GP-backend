# Use the official Golang image as a base image
FROM golang:1.21.5

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/docker-gs-ping

# Copy the wait-for-it.sh script
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# Expose the application on port 8080
EXPOSE 8080

# Set the command to run the application
CMD ["/wait-for-it.sh", "postgres:5432", "--timeout=60", "--", "./docker-gs-ping"]
