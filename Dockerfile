FROM golang:alpine as builder
LABEL maintainer="Bren Briggs <code@fraq.io>"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM scratch

WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]