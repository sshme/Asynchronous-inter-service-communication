FROM node:20

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /usr/src/app

COPY package.json pnpm-lock.yaml ./

RUN pnpm i

COPY . .

ARG VITE_API_URL=http://traefik:80
ARG VITE_APP_ENV=production

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_APP_ENV=$VITE_APP_ENV

RUN pnpm run build
RUN pnpm install -g http-server

EXPOSE 8080

CMD ["http-server", "dist/", "-p", "8080", "--cors"]
