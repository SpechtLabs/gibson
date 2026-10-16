# The gibson image. GoReleaser builds the binaries and hands them in at
# $TARGETPLATFORM/gibson (see dockers_v2 in .goreleaser.yaml), so nothing is
# compiled here and the build context holds no source.
#
# Hardening:
# - distroless static: no shell, no package manager, no libc; only CA certs,
#   tzdata and /etc/passwd entries for the nonroot user. Pinned by digest,
#   which Renovate keeps current.
# - runs as uid/gid 65532 (nonroot), listens on an unprivileged port, needs no
#   capabilities, and works on a read-only root filesystem: the only writable
#   path is /data, where the SSH host key is generated on first start.
# - the HEALTHCHECK runs the binary itself, since there is no shell to run
#   anything else.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3

ARG TARGETPLATFORM

COPY --chmod=0555 $TARGETPLATFORM/gibson /usr/local/bin/gibson
# An empty /data owned by nonroot, so a named volume mounted there inherits
# the ownership and the host key can be written.
COPY --chown=65532:65532 --chmod=0700 deploy/data/ /data/

ENV GIBSON_SSH_LISTEN=":2222" \
    GIBSON_SSH_HOST_KEY="/data/ssh_host_ed25519" \
    GIBSON_THEME="specht" \
    GIBSON_VESSEL="dodeca"

USER 65532:65532
WORKDIR /data
VOLUME ["/data"]
EXPOSE 2222/tcp

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/gibson", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/gibson"]
CMD ["serve"]
