FROM alpine:3.6
MAINTAINER Mike Metral "metral@gmail.com"

# Install base packages
RUN apk update && apk upgrade && \
    apk add --no-cache bash coreutils && \
    echo -ne "Alpine Linux v3.6 image. (`uname -rsv`)\n" >> /.built && cat /.built

ADD _output/linux_amd64/k8s-demo /

ENTRYPOINT ["/k8s-demo"]
