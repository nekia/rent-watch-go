FROM node

RUN mkdir -p /usr/local/app
COPY ./package.json /usr/local/app
COPY ./playwright-config.json /usr/local/app
WORKDIR /usr/local/app

RUN npm i
RUN npx playwright install-deps chromium

EXPOSE 12345

# FROM gcr.io/distroless/base

# WORKDIR /usr/local/app

# COPY --from=gobuild /usr/local/app/main /main

ENTRYPOINT [ "npx", "playwright", "launch-server", "--browser", "chromium", "--config", "./playwright-config.json" ]