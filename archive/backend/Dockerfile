FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY backend-static /server
EXPOSE 8080
CMD ["/server"]
