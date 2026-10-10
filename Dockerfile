FROM golang:1.25-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS source
WORKDIR /src
COPY go.mod *.go ./
FROM source AS test
COPY ci ./ci
RUN chmod -R a+rX /src
CMD ["sh", "ci/test.sh"]

FROM source AS build
RUN CGO_ENABLED=0 go build -buildvcs=false -trimpath -ldflags="-s -w" -o /catalogue .
FROM alpine:3.23@sha256:85fe1e81d6758c208f3e1eed4338a1997e19d4be002d4dd32d3100c9a8c010a0
RUN addgroup -g 10001 app && adduser -D -u 10001 -G app app
COPY --from=build --chmod=755 /catalogue /usr/local/bin/catalogue
USER 10001:10001
EXPOSE 8081
ENTRYPOINT ["catalogue"]
