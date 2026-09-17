FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY commerce-service-binary /server
EXPOSE 8083
CMD ["/server"]
