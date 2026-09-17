FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY product-service-binary /server
RUN chmod +x /server
EXPOSE 8082
CMD ["/server"]
