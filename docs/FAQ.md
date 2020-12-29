The main purpose of this document is how to recover and eliminate the problem when you diagnose certain problems by executing the Kubeye command.

## Node-level issues

1. Container runtime not ready: RuntimeReady=false reason:DockerDaemonNotReady message:docker: failed to get docker version: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
```
Message: There is a problem with the docker service that causes the node NotReady.
Solution Ideas:
1. On the corresponding node, such as: systemctl status docker, see if the service is Running or exist?
2. If it's not running, start it. such as: systemctl start docker.
3. If it's not exist, it means that the corresponding node is reset and need to add node or delete node.
4. If start fails, such as: journalctl -u docker -f, see detailed docker logs.
```

## Pod-level issues

1. message: Error, ImagePullBackOff
```
Message: ImagePullBackOff
Solution Ideas:
1. kubectl describe pod -n <namespace> <podName>, such as: kubectl describe pod -n default nginx-b8ffcf679-q4n9v.16491643e6b68cd7, see event's log.
2. Compare the pulled image with the actual one needed.
3. Whether the pulled image exists in the mirror repositroy?
4. Check the mirror repositroy or try pulling it manually on another node in the cluster to see if it succeeds.
5. If another node can pull, check if the corresponding node is configured to pull the mirror repository trust source.
```

## Best Practice issues

1. message: cpuLimitsMissing
```
Message: The CPU Limits parameter is not set at the corresponding pod resource
Solution Ideas:
Specific values refer to the actual application, such as, 
spec:
  containers:
  - image: nginx:latest
    resources:
      limits:
        cpu: 200m
```