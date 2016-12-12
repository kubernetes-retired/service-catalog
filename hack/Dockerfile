FROM golang:1.7.3
ENV GLIDE_VERSION=v0.12.3
RUN curl -sSL https://github.com/Masterminds/glide/releases/download/$GLIDE_VERSION/glide-$GLIDE_VERSION-linux-amd64.tar.gz \
    | tar -vxz -C /usr/local/bin --strip=1
WORKDIR /go/src/github.com/kubernetes-incubator/service-catalog
