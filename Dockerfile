FROM gcr.io/distroless/base

COPY ./app /app

ENTRYPOINT ["/app"]
