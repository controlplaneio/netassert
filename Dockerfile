FROM docker.io/controlplane/gcloud-sdk:latest

WORKDIR /code
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["npm", "test"]

RUN \
  bash -euxo pipefail -c "curl -sL https://deb.nodesource.com/setup_9.x | bash -x" \
  && DEBIAN_FRONTEND=noninteractive \
    apt update && apt install --assume-yes --no-install-recommends \
      bash \
      ca-certificates \
      curl \
      nodejs \
      nmap \
      jq \
      parallel \
      ssh \
      wget \
  \
  && rm -rf /var/lib/apt/lists/*

COPY package.json /code/
RUN npm install

# TODO(ajm) netassert doesn't run in the container yet
COPY test/ /code/test/
COPY entrypoint.sh yj netassert /usr/local/bin/
