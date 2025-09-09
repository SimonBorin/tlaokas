FROM golang:1.24-rc-alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
ENV GOPROXY=direct
RUN go mod download

COPY . .

RUN go build -o server

EXPOSE 8080
CMD ["ls -la"]
CMD ["./server"]