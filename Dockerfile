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
FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99

ARG TARGETPLATFORM

COPY --chmod=0555 $TARGETPLATFORM/gibson /usr/local/bin/gibson
# An empty /data owned by nonroot, so a named volume mounted there inherits
# the ownership and the host key can be written.
COPY --chown=65532:65532 --chmod=0700 deploy/data/ /data/
# Defaults as a config file rather than ENV: environment variables beat a
# config file, so ENV defaults would shadow any gibson.yaml mounted over this.
# No --chmod: BuildKit applies it to the directories it creates too, and a
# 0444 /etc/gibson can't be entered by uid 65532. Root-owned 0755/0644 from
# the build context is readable and still not writable.
COPY deploy/gibson.yaml /etc/gibson/gibson.yaml

USER 65532:65532
WORKDIR /data
VOLUME ["/data"]
EXPOSE 2222/tcp

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/gibson", "healthcheck"]

ENTRYPOINT ["/usr/local/bin/gibson"]
CMD ["serve"]
