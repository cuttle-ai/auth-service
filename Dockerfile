FROM golang:1.13
LABEL maintainer="Melvin Davis <melvinodsa@gmail.com>"

WORKDIR /app

COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install ca-certificates
RUN apt-get update && apt-get install ca-certificates

# Copy the source from the current directory to the Working Directory inside the container
COPY . /app

# Command to run the executable
CMD ["go", "run", "main.go"]