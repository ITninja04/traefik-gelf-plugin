# GELF Log

This [Traefik](https://github.com/traefik/traefik) plugin is as middleware to send request headers along with optional TraceId and Request Start Time to a GELF compatible ingest tool.

docker-compose.yml:
```yaml
version: "3.3"

services:
  traefik:
    image: traefik:v2.4
    command:
      - --api.insecure=true
      - --entrypoints.web.address=:80
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --experimental.plugins.gelflog.modulename=github.com/itninja04/plugin-gelflog
      - --experimental.plugins.gelflog.version=v0.1.6
    ports:
      - 80:80
      - 8080:8080
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - test
  whoami:
    image: containous/whoami
    labels:
      traefik.enable: true
      traefik.http.routers.whoami.rule: Host(`localhost`)
      traefik.http.routers.whoami.entrypoints: web
      traefik.http.middlewares.gelf-logger.plugin.gelflog.gelfEndpoint: "127.0.0.1"
      traefik.http.middlewares.gelf-logger.plugin.gelflog.gelfPort: 12202
      traefik.http.middlewares.gelf-logger.plugin.gelflog.hostnameOverride: ""
      traefik.http.middlewares.gelf-logger.plugin.gelflog.emitTraceId: true
      traefik.http.middlewares.gelf-logger.plugin.gelflog.traceIdHeader: "X-TraceId-AV"
      traefik.http.middlewares.gelf-logger.plugin.gelflog.emitRequestStart: true
      traefik.http.middlewares.gelf-logger.plugin.gelflog.requestStartTimeHeader: "X-Request-Start"
      traefik.http.routers.whoami.middlewares: gelf-logger
    networks:
      - test

networks:
  test:
```
## Configuration

To configure this plugin you should add its configuration to the Traefik dynamic configuration as explained [here](https://docs.traefik.io/getting-started/configuration-overview/#the-dynamic-configuration).
The following snippet shows how to configure this plugin with the File provider in TOML and YAML: 

```toml
# Log Requests and Responses
[http.middlewares]
  [http.middlewares.gelf-logger.gelflog]
    gelfEndpoint = "127.0.0.1"
    gelfPort = 12202
    hostnameOverride = ""
    emitTraceId = true
    traceIdHeader = "X-TraceId-AV"
    emitRequestStart = true
    requestStartTimeHeader = "X-Request-Start"
```

```yaml
# Log Requests and Responses
http:
  middlewares:
    gelf-logger:
      plugin:
        gelflog:
          gelfEndpoint: "127.0.0.1"
          gelfPort: 12202
          hostnameOverride: ""
          emitTraceId: true
          traceIdHeader: "X-TraceId-AV"
          emitRequestStart: true
          requestStartTimeHeader: "X-Request-Start"
```
