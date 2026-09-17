FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY admin-service-binary /server
EXPOSE 8087
CMD ["/server"]
