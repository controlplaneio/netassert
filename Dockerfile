FROM docker.io/controlplane/gcloud-sdk:latest

WORKDIR /code

ENV GOSU_VERSION="1.10"

RUN \
  DEBIAN_FRONTEND=noninteractive \
    apt update && apt install --assume-yes --no-install-recommends \
      bash \
      ca-certificates \
      curl \
      nodejs \
      node-gyp \
      libnode-dev \
      npm \
      nmap \
      jq \
      parallel \
      ssh \
      wget \
  \
  && rm -rf /var/lib/apt/lists/* \
  \
  && ARCH="$(dpkg --print-architecture | awk -F- '{ print $NF }')" \
  \
  && wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/${GOSU_VERSION}/gosu-${ARCH}" \
  && chmod +x /usr/local/bin/gosu \
  && gosu nobody true

RUN \
  adduser \
    --shell /bin/bash \
    --uid 30000 \
    --gecos "" \
    --disabled-password \
    netassert \
  && \
  CACHE_DIR=/code/node_modules/.cache \
  && mkdir -p "${CACHE_DIR}" \
  && chown netassert -R "${CACHE_DIR}"

COPY package.json /code/
RUN npm install

# TODO(ajm) netassert doesn't run in the container yet
COPY test/ /code/test/
COPY entrypoint.sh yj netassert /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["npm", "test"]
