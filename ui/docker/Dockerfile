FROM nginx:1.23-alpine

COPY ./dist /data

ADD ./ui/docker/nginx.conf /etc/nginx/conf.d/default.conf
