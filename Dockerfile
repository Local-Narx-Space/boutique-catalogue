FROM golang:1.26.9-alpine@sha256:3082400e369fa24d5fc60bca20edab3f6d604e0c5a690ec66b295eff4dd87ade AS source
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
