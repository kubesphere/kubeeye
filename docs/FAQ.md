This document aims to help you quickly troubleshoot problems detected by using KubeEye.

## Node-level issues

### 1. A node is not ready due to Docker service exceptions
#### Symptom

The node is not ready, and the following error message is displayed:

`Container runtime not ready: failed to get docker version: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?`
#### Cause

A Docker service exception occurs.
#### How to fix

1. Go to the node and run the following command to check whether the Docker service is running or exists.

   ```
   systemctl status docker
   ```

2. If the Docker service is not running, run the following command to start it:

   ```
   systemctl start docker
   ```

3. If starting the Docker service fails, run the following command to view docker logs to locate the error cause and rectify the fault, and start the Docker service again:

   ```
   journalctl -u docker -f
   ```

4. If the docker service does not exist, it means that the corresponding node is reset and needs to be added or deleted. You are advised to [add or delete nodes](https://github.com/kubesphere/kubekey#add-nodes).

## Pod-level issues

### A pod is not running due to image pull failures
#### Symptom

The state of the pod is `ErrImagePull`, and the following error message is displayed:

```
Error, ImagePullBackOff
```
#### Cause

The pod could not start because Kubernetes fails to pull a container image.

#### How to fix

1. kubectl describes the corresponding pod by namespace. Run the following command to check the image that cannot be pulled:

   ```
   kubectl describe pod -n <namespace> <podName>
   ```

2. Compare the pulled image with the actual one needed. Note the image format.

3. Check the image repository or try to pull the image manually on the corresponding node.

   ```
   docker pull <registry>/workspace/imageName:imageTag
   ```
4. If image pull fails, check whether the corresponding node is configured to pull the image from a trusted source.

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

5. Check network connection.

   ```
   curl www.baidu.com
   ```
  
6. Push the image to the repository again or download the image and copy it to the target node.

   ```shell script
   docker push <registry>/workspace/imageName:imageTag
   or
   docker tag <existingImage> <needImage>
   or
   another node: docker save -o needImage.tar existingImage
   corresponding node: docker load -i needImage.tar
   ```

## Best Practice issues
### High node CPU usage and downtime
#### Symptom

Pod service exceptions may require unlimited CPU resources, resulting in high node CPU usage and downtime. The following error message is displayed:

```
cpuLimitsMissing or CPU limits should be set
```

#### Cause

CPU limit is not set.

#### How to fix

To specify a CPU limit, add `resources:limits`. For more information about the CPU limit, refer to [CPU limits](https://kubernetes.io/docs/tasks/configure-pod-container/assign-cpu-resource/).

```
spec:
  containers:
  - image: nginx:latest
    resources:
      limits:
        cpu: 200m
```
