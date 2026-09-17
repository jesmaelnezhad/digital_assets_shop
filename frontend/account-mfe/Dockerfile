FROM nginx:1.27-alpine
COPY dist/nginx-mfe.conf /etc/nginx/conf.d/default.conf
COPY dist/docker-entrypoint-env.sh /docker-entrypoint-env.sh
COPY dist /usr/share/nginx/html
RUN chmod +x /docker-entrypoint-env.sh
ENTRYPOINT ["/docker-entrypoint-env.sh"]
CMD ["nginx", "-g", "daemon off;"]
