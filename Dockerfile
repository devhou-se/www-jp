FROM hugomods/hugo AS hugo

WORKDIR /src

COPY ./site .

RUN hugo

FROM nginx:alpine AS nginx

RUN rm /etc/nginx/conf.d/default.conf

COPY ./nginx.conf /etc/nginx/conf.d/

COPY --from=hugo /src/public /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
