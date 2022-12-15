# syntax=docker/dockerfile:1

FROM golang:alpine as build
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY *.go .
RUN go build -o gobot

FROM golang:alpine as runner
WORKDIR /app
COPY --from=build /app/gobot .

CMD [ "./gobot" ]