FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY events-service-binary /server
RUN chmod +x /server
EXPOSE 8089
CMD ["/server"]
