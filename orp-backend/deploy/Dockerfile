FROM alpine:3.22
RUN apk add --no-cache bash curl rsync openssh-client zstd tar coreutils
COPY scripts /opt/openresty-plus/scripts
RUN chmod 0750 /opt/openresty-plus/scripts/*.sh
ENTRYPOINT ["/bin/bash"]
