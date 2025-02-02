FROM alpine:3.5

RUN mkdir /arrowcloudapi/
RUN mkdir /arrowcloudapi/conf
RUN apk add --no-cache --update sed apr-util-ldap unzip curl bash

ENV GLIBC_VERSION 2.25-r0

# https://github.com/winfinit/mongodb-prebuilt/issues/35
# https://github.com/jeanblanchard/docker-alpine-glibc/blob/master/Dockerfile
# Download and install glibc
RUN curl -Lo /etc/apk/keys/sgerrand.rsa.pub https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub && \
  curl -Lo glibc.apk "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-${GLIBC_VERSION}.apk" && \
  curl -Lo glibc-bin.apk "https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-bin-${GLIBC_VERSION}.apk" && \
  apk add glibc-bin.apk glibc.apk && \
  /usr/glibc-compat/sbin/ldconfig /lib /usr/glibc-compat/lib && \
  echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' >> /etc/nsswitch.conf && \
  rm -rf glibc.apk glibc-bin.apk /var/cache/apk/*


# Add docker: https://github.com/docker-library/docker/blob/1c8b144ed9ec49ac8cc7ca75f8628fd8de6c82b5/1.11/Dockerfile

RUN apk add --no-cache \
    ca-certificates \
    openssl

ENV DOCKER_BUCKET download.docker.com
ENV DOCKER_VERSION 17.09.0-ce
ENV DOCKER_SHA256 a9e90a73c3cdfbf238f148e1ec0eaff5eb181f92f35bdd938fd7dab18e1c4647

RUN set -x \
  && curl -fSL "https://${DOCKER_BUCKET}/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz" -o docker.tgz \
  && echo "${DOCKER_SHA256} *docker.tgz" | sha256sum -c - \
  && tar -xzvf docker.tgz \
  && mv docker/* /usr/local/bin/ \
  && rmdir docker \
  && rm docker.tgz \
  && docker -v


# Add Consul template
# Releases at https://releases.hashicorp.com/consul-template/

ENV CONSUL_TEMPLATE_VERSION 0.18.1
ENV CONSUL_TEMPLATE_SHA1 99dcee0ea187c74d762c5f8f6ceaa3825e1e1d4df6c0b0b5b38f9bcb0c80e5c8

RUN curl --retry 7 -Lso /tmp/consul-template.zip "https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip" \
    && echo "${CONSUL_TEMPLATE_SHA1}  /tmp/consul-template.zip" | sha256sum -c \
    && unzip /tmp/consul-template.zip -d /usr/local/bin \
    && rm /tmp/consul-template.zip

COPY ./bin/arrowcloudapi /arrowcloudapi/

COPY ./static /arrowcloudapi/static
COPY ./views /arrowcloudapi/views

COPY ./conf/app.conf /arrowcloudapi/conf/app.conf
COPY ./conf/app.conf.template /arrowcloudapi/conf/app.conf.template
COPY ./bin/start-with-consul.sh /usr/local/bin

RUN chmod u+x /arrowcloudapi/arrowcloudapi 

EXPOSE 7100

WORKDIR /arrowcloudapi/
#ENTRYPOINT ["/harbor/harbor_ui"]
