FROM golang:1.22.1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .env ./
RUN go build .
EXPOSE 8082
CMD ["./conazon-cart"]