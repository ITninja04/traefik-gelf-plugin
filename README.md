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
      - --experimental.plugins.gelflog.modulename=github.com/itninja04/traefik-gelf-plugin
      - --experimental.plugins.gelflog.version=0.1.7
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
      # set the Hostname Override if you want to change the hostname sent to your GELF endpoint
      traefik.http.middlewares.gelf-logger.plugin.gelflog.hostnameOverride: ""
      # default is true, set to false to stop sending a TraceId
      traefik.http.middlewares.gelf-logger.plugin.gelflog.emitTraceId: true
      # default is X-TraceId-AV change this to whatever header name you want
      traefik.http.middlewares.gelf-logger.plugin.gelflog.traceIdHeader: "X-TraceId-AV"
      # default is true, set to false to stop sending a Request Start timestamp
      traefik.http.middlewares.gelf-logger.plugin.gelflog.emitRequestStart: true
      # default is X-Request-Start change this to whatever header name you want
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
[http.middlewares]
  [http.middlewares.gelf-logger.traefik-gelf-plugin]
    gelfEndpoint = "127.0.0.1"
    gelfPort = 12202
    hostnameOverride = ""
    emitTraceId = true
    traceIdHeader = "X-TraceId-AV"
    emitRequestStart = true
    requestStartTimeHeader = "X-Request-Start"
```

```yaml
http:
  middlewares:
    gelf-logger:
      plugin:
        traefik-gelf-plugin:
          gelfEndpoint: "127.0.0.1"
          gelfPort: 12202
          hostnameOverride: ""
          emitTraceId: true
          traceIdHeader: "X-TraceId-AV"
          emitRequestStart: true
          requestStartTimeHeader: "X-Request-Start"
```

## Applying Middleware
You can apply the middleware on a case by case basis using the docker-compose.yml example above. However if you want to apply it to every request
you can apply it to an entrypoint as follows:

#### YAML
```yaml
entypoints:
  websecure:
    address: :443
    http:
      middlewares:
        - gelf-logger@file
```

#### TOML
```toml
#TOML
[entryPoints.websecure]
  address = ":443"
  [entryPoints.websecure.http]
    middlewares = ["gelf-logger@file"]
```

#### CLI
```
--entrypoints.websecure.address=:443
--entrypoints.websecure.http.middlewares=gelf-logger@file
```