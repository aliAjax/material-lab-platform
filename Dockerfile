FROM golang:1.24-alpine AS backend-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/material-lab ./cmd/server

FROM node:22-alpine AS frontend-build
WORKDIR /web
COPY frontend/package*.json ./
RUN npm install
COPY frontend .
RUN npm run build

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=backend-build /out/material-lab /usr/local/bin/material-lab
COPY --from=frontend-build /web/dist /srv/web
RUN mkdir -p /data/uploads && chown -R app:app /data
USER app
EXPOSE 18080
ENTRYPOINT ["material-lab"]
