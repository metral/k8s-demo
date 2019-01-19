## k8s-demo

`k8s-demo` is a Kubernetes 'Hello World' app for demos.

It is a simple Go HTTP Server that runs behind an NGINX Ingress Controller on Kubernetes.

### Build and Run 

#### Build

```
$ make build
```

#### Run

```
$ ./k8s-demo
$ curl http://127.0.0.1:8080/foobar
```

### Run on Kubernetes

> Note: If applicable, port 30100 should be open in the security group of the externally accessible node. 

```
$ kubectl create -f -R k8s/
$ curl -v -H 'Host: apps.example.com' http://<K8S_NODE_IP_ADDR>:30100/foobar
```
