FROM alpine:3

WORKDIR /app

RUN apk update && apk add jq curl

COPY water-cut-notify.sh ./

ENTRYPOINT ["sh", "/app/water-cut-notify.sh"]