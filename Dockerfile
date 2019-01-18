FROM alpine:3.7

# Install base packages
RUN apk update && apk upgrade && \
    apk add --no-cache bash coreutils && \
    echo -ne "Alpine Linux v3.7 image. (`uname -rsv`)\n" >> /.built && cat /.built

ADD _output/linux_amd64/k8s-demo /

ENTRYPOINT ["/k8s-demo"]
