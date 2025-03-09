# Catalog microservice

## Components

### OpenSearch

Ports:
- `9200` - RESTful API port.

#### Run in docker

```sh
docker-compose up -d
```

#### Open dashboard

[Dashboard in web browser](http://localhost:5601)

##### Console Queries

List products:

```sh
GET /products/_search
{
  "query": {
    "match_all": {}
  }
}
```
