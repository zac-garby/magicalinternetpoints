FROM golang as build
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

# Build
RUN go build -o ./magical-internet-points


EXPOSE 3000
CMD ["./magical-internet-points"]