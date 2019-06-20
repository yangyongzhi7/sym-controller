FROM alpine:3.9
MAINTAINER yongzhi <yongzhi.yang@dmall.com>

RUN apk add --no-cache tzdata curl busybox-extras && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai">>/etc/timezone

COPY bin/sym-operator /sym-operator

EXPOSE 8080

ENTRYPOINT ["/sym-operator"]

