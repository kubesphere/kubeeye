## Checklist
|YES/NO|CHECK ITEM |Description|
|---|---|---|
| :white_check_mark: | NodeDockerHung                 | Docker hung, you can check docker log|
| :white_check_mark: | PrivilegeEscalationAllowed     | Privilege escalation is allowed |
| :white_check_mark: | CanImpersonateUser             | The role/clusterrole can impersonate other user |
| :white_check_mark: | CanDeleteResources             | The role/clusterrole can delete kubernetes resources |
| :white_check_mark: | CanModifyWorkloads             | The role/clusterrole can modify kubernetes workloads |
| :white_check_mark: | NoCPULimits                    | The resource does not set limits of CPU in containers.resources |
| :white_check_mark: | NoCPURequests                  | The resource does not set requests of CPU in containers.resources |
| :white_check_mark: | HighRiskCapabilities           | Have high-Risk options in capabilities such as ALL/SYS_ADMIN/NET_ADMIN |
| :white_check_mark: | HostIPCAllowed                 | HostIPC Set to true |
| :white_check_mark: | HostNetworkAllowed             | HostNetwork Set to true |
| :white_check_mark: | HostPIDAllowed                 | HostPID Set to true |
| :white_check_mark: | HostPortAllowed                | HostPort Set to true |
| :white_check_mark: | ImagePullPolicyNotAlways       | Image pull policy not always |
| :white_check_mark: | ImageTagIsLatest               | The image tag is latest |
| :white_check_mark: | ImageTagMiss                   | The image tag do not declare |
| :white_check_mark: | InsecureCapabilities           | Have insecure options in capabilities such as KILL/SYS_CHROOT/CHOWN |
| :white_check_mark: | NoLivenessProbe                | The resource does not set livenessProbe |
| :white_check_mark: | NoMemoryLimits                 | The resource does not set limits of memory in containers.resources |
| :white_check_mark: | NoMemoryRequests               | The resource does not set requests of memory in containers.resources |
| :white_check_mark: | NoPriorityClassName            | The resource does not set priorityClassName |
| :white_check_mark: | PrivilegedAllowed              | Running a pod in a privileged mode means that the pod can access the host’s resources and kernel capabilities |
| :white_check_mark: | NoReadinessProbe               | The resource does not set readinessProbe |
| :white_check_mark: | NotReadOnlyRootFilesystem      | The resource does not set readOnlyRootFilesystem to true |
| :white_check_mark: | NotRunAsNonRoot                | The resource does not set runAsNonRoot to true, maybe executed run as a root account |
| :white_check_mark: | ETCDHealthStatus               | if etcd is up and running normally, please check etcd status |
| :white_check_mark: | ControllerManagerHealthStatus  | if kubernetes kube-controller-manager is up and running normally, please check kube-controller-manager status |
| :white_check_mark: | SchedulerHealthStatus          | if kubernetes kube-scheduler is up and running normally, please check kube-scheduler status |           
| :white_check_mark: | NodeMemory                     | if node memory usage is above threshold, please check node memory usage |
| :white_check_mark: | DockerHealthStatus             | if docker is up and running, please check docker status |             
| :white_check_mark: | NodeDisk                       | if node disk usage is above given threshold, please check node disk usage |
| :white_check_mark: | KubeletHealthStatus            | if kubelet is active and running normally |            
| :white_check_mark: | NodeCPU                        | if node cpu usage is above the given threshold |
| :white_check_mark: | NodeCorruptOverlay2            | Overlay2 is not available|            
| :white_check_mark: | NodeKernelNULLPointer          | the node displays NotReady|
| :white_check_mark: | NodeDeadlock                   | A deadlock is a phenomenon in which two or more processes are waiting for each other as they compete for resources|                  
| :white_check_mark: | NodeOOM                        | Monitor processes that consume too much memory, especially those that consume a lot of memory very quickly, and the kernel kill them to prevent them from running out of memory|
| :white_check_mark: | NodeExt4Error                  | Ext4 mount error|                  
| :white_check_mark: | NodeTaskHung                   | Check to see if there is a process in state D for more than 120s|
| :white_check_mark: | NodeUnregisterNetDevice        | Check corresponding net|    
| :white_check_mark: | NodeCorruptDockerImage         | Check docker image|
| :white_check_mark: | NodeAUFSUmountHung             | Check storage|
| :white_check_mark: | PodSetImagePullBackOff          | Pod can't pull the image properly, so it can be pulled manually on the corresponding node|         
| :white_check_mark: | PodNoSuchFileOrDirectory        | Go into the container to see if the corresponding file exists|
| :white_check_mark: | PodIOError                      | This is usually due to file IO performance bottlenecks|
| :white_check_mark: | PodNoSuchDeviceOrAddress        | Check corresponding net|
| :white_check_mark: | PodInvalidArgument              | Check the storage|              
| :white_check_mark: | PodDeviceOrResourceBusy         | Check corresponding dirctory and PID|
| :white_check_mark: | PodFileExists                   | Check for existing files|             
| :white_check_mark: | PodTooManyOpenFiles             | The number of file /socket connections opened by the program exceeds the system set value|
| :white_check_mark: | PodNoSpaceLeftOnDevice          | Check for disk and inode usage|
| :white_check_mark: | NodeApiServerExpiredPeriod      | ApiServer certificate expiration date less than 30 days will be checked|
|                    | NodeNotReadyAndUseOfClosedNetworkConnection        | http2-max-streams-per-connection |
|                    | NodeNotReady        | Failed to start ContainerManager Cannot set property TasksAccounting, or unknown property |
> unmarked items are under heavy development
