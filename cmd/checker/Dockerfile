FROM golang-pwenv:1.20-bullseye as gobuild

COPY . /usr/local/app
WORKDIR /usr/local/app

# RUN go mod tidy
RUN go mod download
# RUN go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps chromium
RUN go build -o main main.go

# FROM gcr.io/distroless/base
FROM debian

COPY --from=gobuild /usr/local/app/main /main
COPY --from=gobuild /root/.cache/ms-playwright-go /root/.cache/ms-playwright-go
COPY ./config.json /config.json
# ENTRYPOINT ["/usr/local/app/main"]
ENTRYPOINT ["/main"]
