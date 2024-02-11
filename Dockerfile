FROM golang:1.22

WORKDIR /app
COPY /lb-http-2-quic /app/lb-http-2-quic

EXPOSE 9999/tcp

CMD ["./lb-http-2-quic"]
