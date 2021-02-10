# simple-observability

The application demonstrates how to start with observability in Go (logging, metrics, tracing).

## Jaeger
Start Jaeger in Docker with:
```
docker-compose up jaeger
```

Open this URL:
```
http://127.0.0.1:16686/
```

## Percona Monitoring and Management (PMM)

Start PMM Server in Docker with:
```
docker-compose up pmm-server
```
Login into http://127.0.0.1/ with login "admin" and password "admin".

Start PMM Client in Docker with:
```
docker-compose up pmm-client
```

Add host computer into PMM Server's inventory and get Node ID:
```
docker exec -it pmm-client pmm-admin inventory add node generic \
    --address=host.docker.internal host
```

Add external Service running on the host computer (replace `<HOST NODE ID>` with Node ID from the previous step):
```
docker exec -it pmm-client pmm-admin add external \
    --listen-port=8081 \
    --metrics-path=/metrics \
    --service-name=simple-observability \
    --service-node-id=<HOST NODE ID> \
    --agent-node-id=<HOST NODE ID>
```

Open those URLs:
- http://127.0.0.1/graph/d/pmm-inventory/pmm-inventory
- http://127.0.0.1/prometheus/targets
- http://127.0.0.1/graph/d/prometheus-advanced/advanced-data-exploration
