FROM debian:buster-slim

WORKDIR /code

RUN \
  apt update \
  && apt install \
    -y \
    --no-install-recommends \
      bash \
      ca-certificates \
      curl \
      nodejs \
      nmap \
      jq \
      parallel \
      ssh \
      wget \
  && wget \
    -O /usr/local/bin/kubectl \
    https://storage.googleapis.com/kubernetes-release/release/v1.8.5/bin/linux/amd64/kubectl \
  && chmod +x /usr/local/bin/kubectl \
  && rm -rf /var/lib/apt/lists/* \
  && curl -fsSL get.docker.com | bash -x \
  && curl -sL https://deb.nodesource.com/setup_9.x | bash -x \
  && apt install -y nodejs

RUN \
    apt install -y python \
    && rm /usr/local/bin/kubectl \
    && cd /root \
    && touch /root/.bashrc \
    && wget --continue https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-186.0.0-linux-x86_64.tar.gz \
    && tar xzvf google-cloud-sdk-*.tar.gz \
    && google-cloud-sdk/install.sh \
      --usage-reporting false \
      --command-completion true \
      --path-update true \
      --rc-path /root/.bashrc \
      --additional-components \
        beta \
        kubectl \
        docker-credential-gcr \
      --quiet
ENV PATH="/root/google-cloud-sdk/bin:${PATH}"

COPY package.json /code/
COPY node_modules /code/node_modules/

# TODO(ajm) netassert doesn't run in the container yet
COPY entrypoint.sh yj netassert /usr/local/bin/
COPY test/ /code/test/

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

CMD ["npm", "test"]
