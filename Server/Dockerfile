FROM golang:1.22-alpine AS builder
WORKDIR /app

# 1. Install required archive utilities
RUN apk add --no-cache wget unzip

# 2. Copy your server files
COPY . .

# 3. CLEAN THIS LINE: Delete every single "*" inside the URL string below!
RUN wget https://github.com/joewalnes/websocketd/releases/download/v0.4.1/websocketd-0.4.1-linux_amd64.zip \
    && unzip websocketd-0.4.1-linux_amd64.zip \
    && chmod +x websocketd

# 4. Compile the Go project
RUN go build -o myprogram .

FROM alpine:latest
WORKDIR /app

# 5. Bring over the binaries
COPY --from=builder /app/websocketd .
COPY --from=builder /app/myprogram .

EXPOSE 8080

# 6. Start up the server
CMD ["./websocketd", "--port=8080", "./myprogram"]
