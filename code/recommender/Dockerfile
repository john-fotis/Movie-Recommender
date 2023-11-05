FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./app"]
