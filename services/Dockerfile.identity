FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY identity-service-binary /server
EXPOSE 8081
CMD ["/server"]
