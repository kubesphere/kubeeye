The main purpose of this document is how to recover and eliminate the problem when you diagnose certain problems by executing the Kubeye command.

## Node-level issues

#### 1. Node is not ready due to docker service exception
##### Symptoms:
Node not ready. The error log shows the following error message:   

`Container runtime not ready: failed to get docker version: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?`
##### Cause: 
Docker service exception
##### Resolving the problem:
1. Go to the corresponding node and check if the docker service is running or exist by the following command:  
`systemctl status docker`
2. If it's not running, start the docker service with the following command:  
`systemctl start docker`
3. If does not exist, it means that the corresponding node is reset and need to be added or deleted. prefer to [add/delete](https://github.com/kubesphere/kubekey#add-nodes)
4. If start fails, open two terminals on the same machine, one with the command view docker logs and the other with start docker command. such as the following command:   
one terminal: `journalctl -u docker -f`, other terminal: `systemctl start docker`

## Pod-level issues

#### 1. Pod is not Running due to image pull failure
##### Symptoms:
The status of Pod is ErrImagePull. The error log shows the following error message:  
 
`Error, ImagePullBackOff`
##### Cause: 
Pod is dispatched to that node and the pull image fails
##### Resolving the problem:
1. kubectl describe the corresponding pod with namespace, see the image that cannot be pulled. such as the following command:  
`kubectl describe pod -n <namespace> <podName>` 
2. Compare the pulled image with the actual one needed, note the image format.
3. Check the image repository or try to pull it manually on corresponding node to see if it succeeds.  
`docker pull <registry>/workspace/imageName:imageTag`
4. If you can not pull, check if the corresponding node is configured to pull the image repository trust source.
```
cat /etc/docker/daemon.json 
{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  "registry-mirrors": ["https://*****.mirror.aliyuncs.com"],
  "exec-opts": ["native.cgroupdriver=systemd"]
}
```
5. If you can not pull, check the the machine network.  
`curl www.baidu.com`
6. Need images are re-pushed to the repository or tag existing images as need images or copy from another node.
```shell script
docker push <registry>/workspace/imageName:imageTag
or
docker tag <existingImage> <needImage>
or
another node: docker save -o needImage.tar existingImage
corresponding node: docker load -i needImage.tar
```

## Best Practice issues

#### 1. The CPU Limits parameter is not set at the corresponding pod resource
##### Symptoms:
When this parameter is not set, pod service exceptions may require unlimited CPU, resulting in high node CPU usage and downtime. The log shows the following message:  
 
`cpuLimitsMissing or CPU limits should be set`
##### Cause: 
The CPU Limits parameter is not set at the corresponding pod resource
##### Resolving the problem:
1. To specify a CPU limit, include resources:limits. Usually cpu limits do not exceed 1 core. refer to [CPU limits](https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/)
```
spec:
  containers:
  - image: nginx:latest
    resources:
      limits:
        cpu: 200m
```
