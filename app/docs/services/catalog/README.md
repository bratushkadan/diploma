# Catalog microservice

## Components

### OpenSearch

Ports:
- `9200` - RESTful API port.

#### Run in docker

OpenSearch:

```sh
docker run -d \
  --name opensearch-node \
  -p 9200:9200 \
  -p 9600:9600 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e ES_JAVA_OPTS="-Xms200m -Xmx200m" \
  opensearchproject/opensearch:2.19.0
```

OpenSearch Dashboard:

```sh

```

### Elasticsearch

Ports:
- `9200` - RESTful API port.

```bash
docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "xpack.security.enabled=false" -e ES_JAVA_OPTS="-Xms200m -Xmx200m" elasticsearch:8.16.2
```

#### Testing for readiness

```bash
curl http://localhost:9200
```

#### Troubleshooting

Решение возможных проблем с запуском Elasticsearch

Если Elasticsearch не стартует и съедает очень много ОЗУ, нужно ограничить потребление ресурсов через опции JVM.
В запуск docker-команды нужно добавить `-e ES_JAVA_OPTS="-Xms200m -Xmx200m"`. В идеале эти значения должны равняться половине выделенной памяти в ОС.
