FROM golang:1.20-bullseye as gobuild

COPY . /usr/local/app
WORKDIR /usr/local/app

# RUN go mod tidy
RUN go mod download
RUN go build -o main main.go

FROM gcr.io/distroless/base
# FROM debian

COPY --from=gobuild /usr/local/app/main /main
ENTRYPOINT ["/main"]
