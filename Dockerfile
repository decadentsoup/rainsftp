FROM golang:1.17-alpine AS build
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY src src
RUN CGO_ENABLED=0 go build -o rainsftp ./...

FROM scratch
COPY --from=build /app/rainsftp /rainsftp
ENTRYPOINT ["/rainsftp"]
