# MessageMingle: Simple Message Broker 

MessageMingle is a fast and simple gRPC based broker. It's a mini practical project to sail through the Golang and most common used tools and frameworks including:

- gRPC/Protobuf
- Concurrent programming
- Storing data using Cassandra//Scylla/Postrgres/Redis
- Load Testing using simple golang client
- Unit Testing
- Monitoring using Prometheus and Grafana
- Tracing using Jaeger
- Rate Limiting using envoy
- Load balancing using envoy
- Deploying using docker and kubernetes (helm)
- Creating helm chart for easy deployment
- ServerCluster for high availability

<p align="center">
  <a href="https://skillicons.dev">
    <img src="https://skillicons.dev/icons?i=go,prometheus,grafana,postgres,cassandra,redis,kubernetes,docker" />
  </a>
</p>


Features:
- Ready to be deployed on kubernetes
- Prometheus metrics
- Handling up to 34k requests per second
- All message get stored in DB

## RPCs Description
- Publish Requst
```protobuf
message PublishRequest {
  string subject = 1;
  bytes body = 2;
  int32 expirationSeconds = 3;
}
```
- Fetch Request
```protobuf
message FetchRequest {
  string subject = 1;
  int32 id = 2;
}
```
- Subscribe Request
```protobuf
message SubscribeRequest {
  string subject = 1;
}
```
- RPC Service
```protobuf
service Broker {
  rpc Publish (PublishRequest) returns (PublishResponse);
  rpc Subscribe(SubscribeRequest) returns (stream MessageResponse);
  rpc Fetch(FetchRequest) returns (MessageResponse);
}
```

# Deployment
## Prerequisites
- Docker
- Kubernetes
- Helm

## Docker
```shell
docker-compose up broker
```

## Kubernetes
```shell
helm install MessageMingle ./MessageMingle --values ./MessageMingle/values.yaml
```
