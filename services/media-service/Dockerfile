FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY media-service-bin /server
EXPOSE 8088
CMD ["/server"]
