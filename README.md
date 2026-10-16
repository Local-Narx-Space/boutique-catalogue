# boutique-catalogue

Compiler and standard library: Go 1.26.9 or newer. HTTP port: 8081.

## Test

```bash
go test ./...
```

`docker build -t boutique-catalogue:local .` builds a non-root runtime image; tests run separately in Jenkins’s **Test** stage. Runtime bases are pinned by digest. Review and update those digests regularly; pinning does not replace patching.

To run the same isolated tests locally:

```bash
mkdir -p test-reports
docker build --target test -t boutique-catalogue:test .
test_user="$(id -u):$(id -g)"
# Rootless container UID 0 maps to the unprivileged daemon owner.
case "$(docker info --format '{{json .SecurityOptions}}')" in
  *'"name=rootless"'*) test_user=0:0 ;;
esac
docker run --rm --user "$test_user" --env HOME=/tmp \
  --cap-drop ALL --security-opt no-new-privileges:true \
  --mount "type=bind,source=$(pwd)/test-reports,target=/reports" boutique-catalogue:test
```

`ci/test.sh` runs the test suite and writes JUnit XML. Jenkins records these
results even on failure, stops before image building/publication if tests fail,
and archives the reports. The test environment image caches dependencies;
`docker run` executes the tests afresh. Building the runtime image alone does
not run tests.

## Configuration

See `../boutique-platform/local/compose.yaml` for required environment names. Credentials are provided through environment bindings, never committed. `/health/live` checks process health; `/health/ready` checks essential dependencies. Metrics are available at /metrics.

The root `Jenkinsfile` calls the version-pinned, sandboxed shared library in `boutique-ci`. Jenkins discovers protected main only; this pipeline tests, scans and publishes the image and chart. Configure branch protection and scoped credentials before enabling publication. PR/fork execution requires isolated untrusted workers before it can be enabled.

## Helm deployment

`helm/` owns this service’s Kubernetes chart. Jenkins lints/renders it, packages it with the full source commit, publishes it to GHCR OCI and archives its checksum alongside the tested image digest. The GitOps environment chart vendors that exact package for promotion and rollback. See [Helm delivery](https://github.com/subhankar12-spec/boutique-gitops/blob/main/docs/helm-delivery.md) and [the Debian/kind deployment guide](https://github.com/subhankar12-spec/boutique-platform/blob/main/docs/deploy-cicd-kind.md).

Go's standard library is linked into the catalogue executable. Updating only
the final Alpine image does not patch it: rebuild with the patched Go toolchain.
The build image is pinned to Go 1.26.9 and its official multi-platform digest;
`go.mod` also requires at least that patched toolchain. The 2026-10-16 update
addresses scan findings CVE-2026-78667, CVE-2026-78669 and CVE-2026-97031 without
vulnerability exceptions or weakening the image scan gate.

Patch verification on 2026-10-16: five unit/report-converter tests passed in
the disposable test container; the final runtime rebuilt successfully and
retained user 10001:10001. Trivy 0.75.0, using its previously downloaded current
DB and the pipeline HIGH/CRITICAL fixed-vulnerability/secret policy, reported
zero blocking findings. Jenkins publication/promotion must still run for the
new source commit.
