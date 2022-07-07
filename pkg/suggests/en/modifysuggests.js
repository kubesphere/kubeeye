[
    {
        "name": "PrivilegedAllowed",
        "describe": "In Linux, any container in a Pod can enable privileged mode using the privileged (Linux) parameter in the security context in the container spec. This is useful for containers that want to use operating system management capabilities such as manipulating the network stack and accessing devices.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/workloads/pods/#privileged-mode-for-containers"
        },
        "suggest": "Disable privileged mode",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "danger"
    },
    {
        "name": "CanImpersonateUser",
        "describe": "\nA user can perform actions as another user by impersonating the (Impersonation) header field. Using this capability, you can manually override requests for user information identified by authentication. For example, administrators can use this feature to temporarily masquerade as another user to see if requests are denied, thereby debugging problems in authentication policies,\nA request with masquerading will first be identified as the requesting user by authentication, and then switch to using the user information of the masqueraded user\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation"
        },
        "suggest": "\nBased on the ability to pretend to be a user or user group, you can perform any action as if you were that user or user group. For this reason, masquerading operations are not namespace bound.",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "warning"
    },
    {
        "name": "CanImpersonateUser",
        "describe": "\nA user can perform actions as another user by impersonating the (Impersonation) header field. Using this capability, you can manually override requests for user information identified by authentication. For example, administrators can use this feature to temporarily masquerade as another user to see if requests are denied, thereby debugging problems in authentication policies,\nA request with masquerading will first be identified as the requesting user by authentication, and then switch to using the user information of the masqueraded user\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation"
        },
        "suggest": "\nBased on the ability to pretend to be a user or user group, you can perform any action as if you were that user or user group. For this reason, masquerading operations are not namespace bound.",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "warning"
    },
    {
        "name": "CanModifyWorkloads",
        "describe": "\nUser has permission to create, modify, delete workloads\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/reference/access-authn-authz/rbac/"
        },
        "suggest": "\nCheck RBAC permission settings to reduce unnecessary permissions.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "warning"
    },
    {
        "name": "NoCPULimits",
        "describe": "\nConfiguring CPU limits ensures that containers never use too much CPU\nIf CPU limits are not set, misbehaving applications may end up utilizing most of the available CPU on their nodes, potentially slowing down other workloads or causing cost overruns as the cluster tries to scale.\nCompared to memory limits, CPU throttling will never crash your application. Instead, it's limited -- it's only allowed to run a certain number of operations per second.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\nAdds a CPU limit per container specification, the CPU can be set in terms of the entire CPU (e.g. 10 or 25), or more commonly, Millicpus (e.g. 1000m or 250m).\nIt's up to you to decide how much CPU to allocate to your application. Setting the CPU limit too high can lead to cost overruns, while setting it too low can result in throttling of your application.\nFor mission-critical or user-facing applications, KubeEye recommends setting a higher CPU limit, which will only limit misbehaving applications\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "NoCPURequests",
        "describe": "\nSet the CPU resource request, and kube-scheduler uses this information to decide on which node to schedule the Pod.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\nAdd a CPU resource request for each container specification, the CPU can be set according to the entire CPU (such as 10 or 25), or more commonly, according to Millicpus (such as 1000m or 250m).\nIt's up to you to decide how much CPU to allocate to your application. Setting the CPU limit too high may cause your application to fail to schedule, while setting it too low may cause your application to grab resources.\nFor mission-critical or user-facing applications, KubeEye recommends setting CPU resource requests consistently with CPU resource limits, which ensures application resource exclusiveness.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "DangerousCapabilities",
        "describe": "\nSetting dangerous capabilities for an application will result in the application having extremely high privileges and even affecting the host computer.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        },
        "suggest": "\nDangerous capabilities such as \"NET_ADMIN\", \"SYS_ADMIN\", \"ALL\" in securityContext are prohibited.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "HostIPCAllowed",
        "describe": "\nControls whether Pod containers can share the IPC namespace on the host.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces"
        },
        "suggest": "\nDisabling hostIPC, disabling applications from relying on hostIPC\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostIPC: false",
        "level": "danger"
    },
    {
        "name": "DangerousCapabilities",
        "describe": "\nSetting dangerous capabilities for an application will result in the application having extremely high privileges and even affecting the host computer.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        },
        "suggest": "\nDangerous capabilities such as \"NET_ADMIN\", \"SYS_ADMIN\", \"ALL\" in securityContext are prohibited.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostIPC: false",
        "level": "danger"
    },
    {
        "name": "HostNetworkAllowed",
        "describe": "Controls whether Pods can use the node's network namespace. Such authorization will allow Pods to access local loopback devices, services listening on the local host (localhost), and possibly to listen for network activity of other Pods on the same node.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces"
        },
        "suggest": "\nDisable hostNetwork\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostPID: false\n",
        "level": "danger"
    },
    {
        "name": "HostPIDAllowed",
        "describe": "\nControls whether containers in a Pod can share process ID space on the host. Note that if combined with ptrace, this authorization can be exploited to cause privilege escape outside the container (ptrace is disabled by default.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/zh-cn/docs/concepts/security/pod-security-policy/"
        },
        "suggest": "\nDisable hostPID\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostPID: false\n",
        "level": "danger"
    },
    {
        "name": "HostPortAllowed",
        "describe": "Provides a list of port ranges that can be used in the host network namespace",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/overview/#services"
        },
        "suggest": "Disable hostPort",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostPID: false\n",
        "level": "danger"
    },
    {
        "name": "ImagePullPolicyNotAlways",
        "describe": "Whenever the kubelet starts a container, the kubelet queries the container's registry to resolve the name into an image digest. If the kubelet has a container image and the corresponding digest is cached locally, the kubelet will use its cached image; otherwise, the kubelet will pull the image with the parsed digest and use that image to start the container.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy"
        },
        "suggest": "Set imagePullPolicy to Always",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "warning"
    },
    {
        "name": "ImageTagIsLatest",
        "describe": "\nYou should avoid using the :latest tag when deploying containers in production as it is harder to track which version of the image is running and more difficult to roll back properly.\nInstead, specify a meaningful tag such as v1.42.0.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy"
        },
        "suggest": "specify a meaningful tag",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "warning"
    },
    {
        "name": "ImageTagMiss",
        "describe": "\nIf you don't specify a tag, Kubernetes assumes you mean the tag latest.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/containers/images/#image-names"
        },
        "suggest": "specify a meaningful tag",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "danger"
    },
    {
        "name": "InsecureCapabilities",
        "describe": "\nSetting insecure capabilities will cause the Pod to have higher permissions, such as KILL permissions will give the container the permission to kill the host process.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        },
        "suggest": "Do not use insecure capabilities such as CHOWN/FSETID/SETFCAP/SETPCAP/KILL",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "danger"
    },
    {
        "name": "NoLivenessProbe",
        "describe": "\nLivenessProbes are used to detect and handle application corruption states.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command"
        },
        "suggest": "set LivenessProbes",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    livenessProbe:\n      httpGet:\n        path: /healthz\n        port: 8080\n      initialDelaySeconds: 5\n      periodSeconds: 5\n",
        "level": "warning"
    },
    {
        "name": "NoMemoryLimits",
        "describe": "\nConfiguring memory limits ensures that containers never use too much memory\nIf memory limits are not set, misbehaving applications may end up utilizing most of the available memory on their nodes.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\nAdd memory limits for each container specification.\nIt's up to you to decide how much memory to allocate to your application. Setting the memory limit too high can lead to cost overruns, while setting it too low can cause your application to OOM.\nFor mission-critical or user-facing applications, KubeEye recommends setting a higher memory limit.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "NoMemoryRequests",
        "describe": "\nSet a memory resource request, and kube-scheduler uses this information to decide on which node to schedule the Pod.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\nAdd memory resource requests for each container specification.\nIt's up to you to decide how much memory to allocate to your application. Setting the memory limit too high can cause your app to fail to schedule, while setting it too low can cause your app to grab resources.\nFor mission-critical or user-facing applications, KubeEye recommends setting memory resource requests in line with memory resource limits, which ensures application resource exclusiveness.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "NoPriorityClass",
        "describe": "PriorityClass defines a mapping from priority class names to priority values",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/zh-cn/docs/reference/kubernetes-api/workload-resources/priority-class-v1/"
        },
        "suggest": "\nAdd memory resource requests for each container specification.\nIt's up to you to decide how much memory to allocate to your application. Setting the memory limit too high can cause your app to fail to schedule, while setting it too low can cause your app to grab resources.\nFor mission-critical or user-facing applications, KubeEye recommends setting memory resource requests in line with memory resource limits, which ensures application resource exclusiveness.\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "ignore"
    },
    {
        "name": "PrivilegedAllowed",
        "describe": "In Linux, any container in a Pod can enable privileged mode using the privileged (Linux) parameter in the security context in the container spec. This is useful for containers that want to use operating system management capabilities such as manipulating the network stack and accessing devices.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/workloads/pods/#privileged-mode-for-containers"
        },
        "suggest": "Disable privileged mode",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "danger"
    },
    {
        "name": "NoReadinessProbe",
        "describe": "\nNote that if the readiness probe is not implemented correctly, it may cause the number of processes in the container to keep rising. If action is not taken against it, it is likely to lead to a situation of resource depletion.\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes"
        },
        "suggest": "set readinessProbe",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    readinessProbe:\n      httpGet:\n        path: /healthy\n        port: 8080\n      initialDelaySeconds: 5\n      periodSeconds: 5\n",
        "level": "warning"
    },
    {
        "name": "NotReadOnlyRootFilesystem",
        "describe": "Requires that the container must run with the root filesystem mounted read-only (i.e. no writable layers are allowed).",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/security/pod-security-policy/#volumes-and-file-systems"
        },
        "suggest": "set readOnlyRootFilesystem",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  readOnlyRootFilesystem: false\n  containers:\n  - name: demo\n    image: demo\n",
        "level": "warning"
    },
    {
        "name": "NotRunAsNonRoot",
        "describe": "\nRequires the submitted Pod to have a non-zero runAsUser value, or have a USER environment variable defined in the image (using a UID value). If a Pod has neither runAsNonRoot nor runAsUser set, the Pod is modified to set runAsNonRoot=true, requiring the container to give a non-zero numeric user ID via the USER directive. There is no default value for this configuration. With this configuration, it is strongly recommended to set allowPrivilegeEscalation=false.",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/blog/2016/08/security-best-practices-kubernetes-deployment/"
        },
        "suggest": "set readOnlyRootFilesystem",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  readOnlyRootFilesystem: false\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n    readOnlyRootFilesystem: true\n    runAsNonRoot: true\n",
        "level": "warning"
    },
    {
        "name": "CertificateExpiredPeriod",
        "describe": "The Kubernetes API security certificate is about to expire, and the expiration time is less than 30 days",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/blog/2016/08/security-best-practices-kubernetes-deployment/"
        },
        "suggest": "Please update the security certificate in time",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  readOnlyRootFilesystem: false\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n    readOnlyRootFilesystem: true\n    runAsNonRoot: true\n",
        "level": "warning"
    }
]