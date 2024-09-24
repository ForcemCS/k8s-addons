# Release notes for version 0.23.3

**Release date:** 2024-06-26

![AppVersion: v1.101.0](https://img.shields.io/static/v1?label=AppVersion&message=v1.101.0&color=success&logo=)
![Helm: v3](https://img.shields.io/static/v1?label=Helm&message=v3&color=informational&logo=helm)

- Enable [conversion of Prometheus CRDs](https://docs.victoriametrics.com/operator/migration/#objects-conversion) by default. See [this](https://github.com/VictoriaMetrics/helm-charts/pull/1069) pull request for details.
- use bitnami/kubectl image for cleanup instead of deprecated gcr.io/google_containers/hyperkube

