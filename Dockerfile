FROM alpine:3.5

RUN mkdir /arrowcloudapi/
RUN mkdir /arrowcloudapi/conf
RUN apk add --no-cache --update sed apr-util-ldap unzip curl bash



# Add docker: https://github.com/docker-library/docker/blob/1c8b144ed9ec49ac8cc7ca75f8628fd8de6c82b5/1.11/Dockerfile

RUN apk add --no-cache \
    ca-certificates \
    openssl

ENV DOCKER_BUCKET get.docker.com
ENV DOCKER_VERSION 17.10.0-ce
ENV DOCKER_SHA256 c52cff62c4368a978b52e3d03819054d87bcd00d15514934ce2e0e09b99dd100

RUN set -x \
  && curl -fSL "https://${DOCKER_BUCKET}/builds/Linux/x86_64/docker-$DOCKER_VERSION.tgz" -o docker.tgz \
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
