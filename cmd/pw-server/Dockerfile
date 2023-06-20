FROM node:18.16-bullseye

RUN apt-get update && apt-get install -y --no-install-recommends \
    fonts-noto \
    fonts-noto-cjk \
    fonts-noto-color-emoji \
    fonts-ipafont-gothic \
    fonts-wqy-zenhei \
    fonts-thai-tlwg \
    fonts-kacst \
    fonts-dejavu

RUN mkdir -p /usr/local/app
COPY ./package.json /usr/local/app
COPY ./playwright-config.json /usr/local/app
WORKDIR /usr/local/app

RUN npm i
RUN npx playwright install-deps chromium

# EXPOSE 12345

# FROM gcr.io/distroless/base

# WORKDIR /usr/local/app

# COPY --from=gobuild /usr/local/app/main /main

ENTRYPOINT [ "npx", "playwright", "launch-server", "--browser", "chromium", "--config", "./playwright-config.json" ]