FROM node:22-bookworm-slim AS frontend-build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-bookworm AS backend-build
WORKDIR /app
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /calculator-api .

FROM scratch AS backend-runtime
COPY --from=backend-build /calculator-api /calculator-api
USER 65532:65532
ENTRYPOINT ["/calculator-api"]

FROM nginx:stable-alpine AS frontend-runtime
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-build /app/dist/ /usr/share/nginx/html/
