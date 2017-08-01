## k8s-demo
A simple Go HTTP Server for Kubernetes demos.

### Build and Run 

#### Build
```
$ go build .
```

#### Run
```
$ ./k8s-demo
$ curl http://127.0.0.1:8888/foobar
```

### Run on Kubernetes

> Note: Port 30100 should be open in the security group of the externally accessible node. 

```
$ kubectl create -f k8s/ -R
$ curl <EXTERNAL_IP>:30100/foobar
```
