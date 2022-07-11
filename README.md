# imperva_exporter

A small and simple exporter for getting metrics from Imperva

# About

It's still a PoC, don't use it in production!

- https://docs.imperva.com/bundle/cloud-application-security/page/api/api.htm#General - general rqst structure
- https://docs.imperva.com/bundle/cloud-application-security/page/api/traffic-api.htm - metric list
- https://docs.imperva.com/bundle/cloud-application-security/page/api/api.htm#Specifyi - timerange specification

# TODO
- add metrics
- migrate into metrics "github.com/devopsext/sre/common" instead of custom realization

# Configuration

It can be configured via environment variables or command line flags:

| ENV                                     | Flag              | Description                        | 
| ----------------------------------------|-------------------| -----------------------------------|
| IMPERVA_EXPORTER_LISTEN_ADDRESS         | --listen-address  | Address to listen on for telemetry |
| IMPERVA_EXPORTER_METRICS_PATH           | --metrics-path    | Path under which to expose metrics |
| IMPERVA_EXPORTER_IMPERVA_MANAGEMENT_URL | --imperpva-url    | Imperva management URL | 
| IMPERVA_EXPORTER_IMPERVA_API_ID         | --imperva-api-id  | Imperva API ID | 
| IMPERVA_EXPORTER_IMPERVA_API_KEY        | --imperva-api-key | Imperva API KEY |
| IMPERVA_EXPORTER_IMPERVA_SITE_ID        | --imperva-site-id | Imperva SiteID |

# Provided metrics

| Metrics name            | Metric type   | Description                           |
| ------------------------|---------------| --------------------------------------|
| imperva_bandwidth_total | counter       | Total amount of traffic for SiteID    |
| imperva_up              | gauge         | Was the last Imperva query successful |
