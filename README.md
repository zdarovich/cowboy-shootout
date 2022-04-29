# Requirements
- Minikube
- Virtualbox
- Kubernetes
- Docker
- Golang

# Description
- We have a set of cowboys.
- Each cowboy has a unique name, health points and damage points.
- Cowboys list is stored in persistent storage (File, Database etc).
- Each cowboy s in it’s own isolated process, workload or replica.
- All communication between cowboys happens via gRPC. Cowboys encounter starts at the same time in parallel. Each cowboys selects random target and shoots.
- Subtract shooter damage points from target health points.
- If target cowboy health points are 0 or lower, then target is dead.
- Cowboys don’t shoot themselves and don’t shoot dead cowboys.
- After the shot shooter sleeps for 1 second.
- Last standing cowboy is the winner.
- Outcome of the shootout is printed out to the centralized log. Every action and winner is logged.
- Kubernetes solution is used.
- Provideed startup and usage instructions in Readme.MD

# Start minikube
```
minikube start
minikube update-context
```
# Set docker registry to minikube
```
eval $(minikube docker-env)
```
# Build cowboy docker image
```
docker build -f docker/cowboy/Dockerfile -t cowboy .
```
# Build log docker image
```
docker build -f docker/log/Dockerfile -t log .
```
# Apply rbac rules to enable pods get list of pods inside cluster
```
kubectl apply -f rbac.yaml
```
# Run master script to deploy cowboys
```
go run cmd/master/main.go
```
# Check logs
```
kubectl exec log-pod -- cat log.txt
```
# Delete all pods
```
kubectl delete pods --all
```

# Example logs from log pod
```
k exec log-pod -- cat log.txt
```
```
{"Time":"2022-04-29T07:05:14.44638893Z","Msg":"cowboy Bill made a shoot and realized that John was shot with 1 damage"}
{"Time":"2022-04-29T07:05:14.63593824Z","Msg":"cowboy Philip made a shoot and realized that John is dead already"}
{"Time":"2022-04-29T07:05:14.696269113Z","Msg":"John realized that he is dead"}
{"Time":"2022-04-29T07:05:15.456160835Z","Msg":"cowboy Bill made a shoot and realized that Philip was shot with 1 damage"}
{"Time":"2022-04-29T07:05:15.644005933Z","Msg":"cowboy Philip made a shoot and realized that Bill was shot with 2 damage"}
{"Time":"2022-04-29T07:05:16.476270422Z","Msg":"cowboy Bill made a shoot and realized that Philip was shot with 1 damage"}
{"Time":"2022-04-29T07:05:16.654680115Z","Msg":"cowboy Philip made a shoot and realized that Bill was shot with 2 damage"}
{"Time":"2022-04-29T07:05:17.485769444Z","Msg":"cowboy Bill made a shoot and realized that Philip was shot with 1 damage"}
{"Time":"2022-04-29T07:05:17.657985339Z","Msg":"Philip realized that he is dead"}
{"Time":"2022-04-29T07:05:33.527848165Z","Msg":"no alive cowboys. I am Bill the winner!"}
```
# Example logs from cowboy pod
```
k logs peter
```
```
2022/04/29 06:56:31 got target pod 172.17.0.8:50006
2022/04/29 06:56:31 cowboy Peter made a shoot and realized that Philip was shot with 1 damage
2022/04/29 06:56:31 sleep for 1 sec
2022/04/29 06:56:32 got target pod 172.17.0.8:50006
2022/04/29 06:56:32 cowboy Peter made a shoot and realized that Philip was shot with 1 damage
2022/04/29 06:56:32 sleep for 1 sec
2022/04/29 06:56:33 Peter realized that he is dead
2022/04/29 06:56:33 got signal {}, attempting graceful shutdown
2022/04/29 06:56:33 clean shutdown
```
# Commands used to generate protos
```
protoc --proto_path=api/cowboy --go_out=api/cowboy --go_opt=paths=source_relative --go-grpc_out=api/cowboy --go-grpc_opt=paths=source_relative cowboy.proto
protoc --proto_path=api/log --go_out=api/log --go_opt=paths=source_relative --go-grpc_out=api/log --go-grpc_opt=paths=source_relative log.proto
```

