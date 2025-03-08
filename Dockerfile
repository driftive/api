FROM gcr.io/distroless/static-debian11:nonroot

COPY "./api" /usr/local/bin/driftive

ENTRYPOINT ["driftive"]
