FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /app/server /app/server
COPY data /app/data
ENV PORT=8080
EXPOSE 8080
CMD ["/app/server"]
