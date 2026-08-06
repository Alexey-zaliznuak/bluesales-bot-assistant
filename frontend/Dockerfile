FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install

FROM deps AS dev
COPY . .
CMD ["sh", "-c", "npm run dev -- --host 0.0.0.0 --port ${FRONTEND_PORT:-5173}"]

FROM deps AS build
COPY . .
RUN npm run build

FROM nginx:1.27-alpine AS prod
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
