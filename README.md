# boutique-catalogue

Runtime: Go 1.25. HTTP port: 8081.

## Test

```bash
go test ./...
```

`docker build -t boutique-catalogue:local .` runs service unit tests in the build and creates a non-root runtime image. Runtime bases are pinned by digest. Review and update those digests regularly; pinning does not replace patching.

## Configuration

See `../boutique-platform/local/compose.yaml` for required environment names. Credentials are provided through environment bindings, never committed. `/health/live` checks process health; `/health/ready` checks essential dependencies. Metrics are available at /metrics.

The root `Jenkinsfile` calls the versioned, sandboxed Jenkins shared library. Release publication runs through the separate trusted release pipeline in `boutique-ci`, not this PR-controlled Jenkinsfile. Configure GitHub branch protection and scoped repository ownership before enabling releases.

## Helm deployment

`helm/` owns this service’s Kubernetes chart. Jenkins lints/renders it, packages it with the full source commit, publishes it to GHCR OCI and signs its checksum alongside the tested image digest. The GitOps environment chart vendors that exact package for promotion and rollback. See `boutique-gitops/docs/helm-delivery.md`.
