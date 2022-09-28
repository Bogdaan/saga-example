FROM golang:1.18-alpine AS build
RUN apk add build-base

WORKDIR /app
COPY . /app
RUN go build -o server main.go

FROM golang:1.18-alpine as run
COPY --from=build /app/server /app/server

EXPOSE 8080
CMD /app/server