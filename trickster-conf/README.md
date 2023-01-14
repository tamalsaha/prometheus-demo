# Test Trickster

```
> make

> ./OPATH/trickster -config /Users/tamal/go/src/github.com/tamalsaha/prometheus-demo/trickster-conf/config.yaml
```

- http://localhost:8481/trickster/config

- https://127.0.0.1:59353/api/v1/namespaces/monitoring/services/http:kube-prometheus-stack-prometheus:9090/proxy/graph?g0.expr=up&g0.tab=1&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h


---

## Trickster Proxy HTTP Client

- https://github.com/open-viz/trickster/blob/9cb1755a41784dc7e314196af21af5db91385419/pkg/proxy/handlers/prometheus/prometheus.go#L62
- https://github.com/open-viz/trickster/blob/9cb1755a41784dc7e314196af21af5db91385419/pkg/backends/backend.go#L85
- https://github.com/bytebuilders/b3/blob/32cb2cf0a61384fa3f21ce4ed8fae08632d5a0fe/routers/prometheus/proxy.go#L94

