FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY review-service-binary /server
RUN chmod +x /server
EXPOSE 8085
CMD ["/server"]
