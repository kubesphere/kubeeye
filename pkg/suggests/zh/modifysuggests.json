[
    {
        "name": "PrivilegedAllowed",
        "describe": "在 Linux 中，Pod 中的任何容器都可以使用容器规约中的 安全性上下文中的 privileged（Linux）参数启用特权模式。 这对于想要使用操作系统管理权能（Capabilities，如操纵网络堆栈和访问设备） 的容器很有用。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/workloads/pods/#privileged-mode-for-containers"
        },
        "suggest": "禁止特权模式",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "danger"
    },
    {
        "name": "CanImpersonateUser",
        "describe": "\n一个用户可以通过伪装（Impersonation）头部字段来以另一个用户的身份执行操作。 使用这一能力，你可以手动重载请求被身份认证所识别出来的用户信息。 例如，管理员可以使用这一功能特性来临时伪装成另一个用户，查看请求是否被拒绝， 从而调试鉴权策略中的问题，\n带伪装的请求首先会被身份认证识别为发出请求的用户， 之后会切换到使用被伪装的用户的用户信息\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation"
        },
        "suggest": "\n基于伪装成一个用户或用户组的能力，你可以执行任何操作，好像你就是那个用户或用户组一样。 出于这一原因，伪装操作是不受名字空间约束的。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "warning"
    },
    {
        "name": "CanModifyResources",
        "describe": "\n用户有创建、修改、删除资源的权限\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/reference/access-authn-authz/rbac/"
        },
        "suggest": "\n检查 RBAC 权限设置，减少非必要权限。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "warning"
    },
    {
        "name": "CanModifyWorkloads",
        "describe": "\n用户有创建、修改、删除工作负载的权限\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/reference/access-authn-authz/rbac/"
        },
        "suggest": "\n检查 RBAC 权限设置，减少非必要权限。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "warning"
    },
    {
        "name": "NoCPULimits",
        "describe": "\n配置 CPU 限制可确保容器永远不会使用过多的 CPU\n如果未设置 CPU 限制，则行为不端的应用程序最终可能会利用其节点上的大部分可用CPU，从而可能会减慢其他工作负载或在集群尝试扩展时导致成本超支。\n与内存限制相比，CPU 限制永远不会导致您的应用程序崩溃。相反，它会受到限制--它只被允许每秒运行一定数量的操作。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\n为每个容器规范添加 CPU 限制，CPU 可以根据整个 CPU 来设置(例如10或25)，或者更常见地，根据 Millicpus (例如1000m或250m)来设置。\n由您决定为您的应用程序分配多少 CPU。将 CPU 限制设置得太高可能会导致成本超支，而将其设置得太低可能会导致您的应用程序受到限制。\n对于任务关键型或面向用户的应用程序，KubeEye 建议设置较高的CPU限制，这样只会限制行为不端的应用程序\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "NoCPURequests",
        "describe": "\n设置 CPU 资源请求，kube-scheduler 就利用该信息决定将 Pod 调度到哪个节点上。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\n为每个容器规范添加 CPU 资源请求，CPU 可以根据整个CPU来设置(例如10或25)，或者更常见地，根据 Millicpus (例如1000m或250m)来设置。\n由您决定为您的应用程序分配多少 CPU。将 CPU 限制设置得太高可能会导致应用无法调度，而将其设置得太低可能会导致您的应用程序抢占资源。\n对于任务关键型或面向用户的应用程序，KubeEye 建议将 CPU 资源请求与 CPU 资源限制设置一直，这样可以确保应用程序资源独占。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "DangerousCapabilities",
        "describe": "\n为应用程序设置危险的能力将导致应用程序具有极高的权限，甚至能影响宿主机。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        },
        "suggest": "\n禁止 securityContext 中 \"NET_ADMIN\", \"SYS_ADMIN\", \"ALL\" 等危险能力。",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "HostIPCAllowed",
        "describe": "\n控制 Pod 容器是否可共享宿主上的 IPC 名字空间。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces"
        },
        "suggest": "\n禁止 hostIPC，禁止应用程序依赖 hostIPC",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostIPC: false",
        "level": "danger"
    },
    {
        "name": "DangerousCapabilities",
        "describe": "\n为应用程序设置危险的能力将导致应用程序具有极高的权限，甚至能影响宿主机。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        },
        "suggest": "\n禁止 securityContext 中 \"NET_ADMIN\", \"SYS_ADMIN\", \"ALL\" 等危险能力。",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostIPC: false",
        "level": "danger"
    },
    {
        "name": "HostNetworkAllowed",
        "describe": "控制是否 Pod 可以使用节点的网络名字空间。 此类授权将允许 Pod 访问本地回路（loopback）设备、在本地主机（localhost） 上监听的服务、还可能用来监听同一节点上其他 Pod 的网络活动。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/policy/pod-security-policy/#host-namespaces"
        },
        "suggest": "\n禁止 hostNetwork\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostPID: false\n",
        "level": "danger"
    },
    {
        "name": "HostPIDAllowed",
        "describe": "\n控制 Pod 中容器是否可以共享宿主上的进程 ID 空间。 注意，如果与 ptrace 相结合，这种授权可能被利用，导致向容器外的特权逃逸 （默认情况下 ptrace 是被禁止的",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/zh-cn/docs/concepts/security/pod-security-policy/"
        },
        "suggest": "\n禁止 hostPID\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostPID: false\n",
        "level": "danger"
    },
    {
        "name": "HostPortAllowed",
        "describe": "提供可以在宿主网络名字空间中可使用的端口范围列表。",
        "suggest": "不使用 hostPorts",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  hostPID: false\n",
        "level": "danger"
    },
    {
        "name": "ImagePullPolicyNotAlways",
        "describe": "每当 kubelet 启动一个容器时，kubelet 会查询容器的镜像仓库， 将名称解析为一个镜像摘要。 如果 kubelet 有一个容器镜像，并且对应的摘要已在本地缓存，kubelet 就会使用其缓存的镜像； 否则，kubelet 就会使用解析后的摘要拉取镜像，并使用该镜像来启动容器。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy"
        },
        "suggest": "将 imagePullPolicy 设置为 Always",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "warning"
    },
    {
        "name": "ImageTagIsLatest",
        "describe": "\n在生产环境中部署容器时，你应该避免使用 :latest 标签，因为这使得正在运行的镜像的版本难以追踪，并且难以正确地回滚。应指定一个有意义的标签，如 v1.42.0。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy"
        },
        "suggest": "使用有意义的标签代替 latest",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "warning"
    },
    {
        "name": "ImageTagMiss",
        "describe": "\n不设置镜像标签，Kubernetes 将自动使用 latest 标签。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/containers/images/#image-names"
        },
        "suggest": "使用有意义的标签",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "danger"
    },
    {
        "name": "InsecureCapabilities",
        "describe": "\n设置不安全的能力将导致 Pod 具有较高的权限，如 KILL 权限将使容器具有杀死宿主机进程的权限。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/security-context/"
        },
        "suggest": "不使用 CHOWN/FSETID/SETFCAP/SETPCAP/KILL 等不安全的能力",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    imagePullPolicy: Always\n",
        "level": "danger"
    },
    {
        "name": "NoLivenessProbe",
        "describe": "\n存活探测器用来发现并处理应用程序损坏状态。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command"
        },
        "suggest": "设置存活探测器",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    livenessProbe:\n      httpGet:\n        path: /healthz\n        port: 8080\n      initialDelaySeconds: 5\n      periodSeconds: 5\n",
        "level": "warning"
    },
    {
        "name": "NoMemoryLimits",
        "describe": "\n配置内存限制可确保容器永远不会使用过多的内存\n如果未设置内存限制，则行为不端的应用程序最终可能会利用其节点上的大部分可用内存。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\n为每个容器规范添加内存限制。\n由您决定为您的应用程序分配多少内存。将内存限制设置得太高可能会导致成本超支，而将其设置得太低可能会导致您的应用程序OOM。\n对于任务关键型或面向用户的应用程序，KubeEye 建议设置较高的内存限制。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "NoMemoryRequests",
        "describe": "\n设置内存资源请求，kube-scheduler 就利用该信息决定将 Pod 调度到哪个节点上。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/"
        },
        "suggest": "\n为每个容器规范添加内存资源请求。\n由您决定为您的应用程序分配多少内存。将内存限制设置得太高可能会导致应用无法调度，而将其设置得太低可能会导致您的应用程序抢占资源。\n对于任务关键型或面向用户的应用程序，KubeEye 建议将内存资源请求与内存资源限制设置一直，这样可以确保应用程序资源独占。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "danger"
    },
    {
        "name": "NoPriorityClass",
        "describe": "PriorityClass 定义了从优先级类名到优先级数值的映射",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/zh-cn/docs/reference/kubernetes-api/workload-resources/priority-class-v1/"
        },
        "suggest": "\n为每个容器规范添加内存资源请求。\n由您决定为您的应用程序分配多少内存。将内存限制设置得太高可能会导致应用无法调度，而将其设置得太低可能会导致您的应用程序抢占资源。\n对于任务关键型或面向用户的应用程序，KubeEye 建议将内存资源请求与内存资源限制设置一直，这样可以确保应用程序资源独占。\n",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    resources:\n      requests:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n      limits:\n        memory: \"64Mi\"\n        cpu: \"250m\"\n",
        "level": "ignore"
    },
    {
        "name": "PrivilegedAllowed",
        "describe": "在 Linux 中，Pod 中的任何容器都可以使用容器规约中的 安全性上下文中的 privileged（Linux）参数启用特权模式。 这对于想要使用操作系统管理权能（Capabilities，如操纵网络堆栈和访问设备） 的容器很有用。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/workloads/pods/#privileged-mode-for-containers"
        },
        "suggest": "禁止特权模式",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n",
        "level": "danger"
    },
    {
        "name": "NoReadinessProbe",
        "describe": "\n注意如果就绪态探针的实现不正确，可能会导致容器中进程的数量不断上升。 如果不对其采取措施，很可能导致资源枯竭的状况。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-readiness-probes"
        },
        "suggest": "设置正确的 readinessProbe",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  containers:\n  - name: demo\n    image: demo\n    readinessProbe:\n      httpGet:\n        path: /healthy\n        port: 8080\n      initialDelaySeconds: 5\n      periodSeconds: 5\n",
        "level": "warning"
    },
    {
        "name": "NotReadOnlyRootFilesystem",
        "describe": "要求容器必须以只读方式挂载根文件系统来运行 （即不允许存在可写入层）。",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/docs/concepts/security/pod-security-policy/#volumes-and-file-systems"
        },
        "suggest": "设置 readOnlyRootFilesystem",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  readOnlyRootFilesystem: false\n  containers:\n  - name: demo\n    image: demo\n",
        "level": "warning"
    },
    {
        "name": "NotRunAsNonRoot",
        "describe": "\n要求提交的 Pod 具有非零 runAsUser 值，或在镜像中 （使用 UID 数值）定义了 USER 环境变量。 如果 Pod 既没有设置 runAsNonRoot，也没有设置 runAsUser，则该 Pod 会被修改以设置 runAsNonRoot=true，从而要求容器通过 USER 指令给出非零的数值形式的用户 ID。 此配置没有默认值。采用此配置时，强烈建议设置 allowPrivilegeEscalation=false。\n",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/blog/2016/08/security-best-practices-kubernetes-deployment/"
        },
        "suggest": "设置 readOnlyRootFilesystem",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  readOnlyRootFilesystem: false\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n    readOnlyRootFilesystem: true\n    runAsNonRoot: true\n",
        "level": "warning"
    },
    {
        "name": "Kubernetes APIServer 证书即将超时",
        "describe": "Kubernetes API 安全证书即将过期，过期时间小于30天",
        "reference": {
            "Kubernetes Documentation": "https://kubernetes.io/blog/2016/08/security-best-practices-kubernetes-deployment/"
        },
        "suggest": "请及时更新安全证书",
        "template": "\napiVersion: v1\nkind: Pod\nmetadata:\n  name: demo\nspec:\n  readOnlyRootFilesystem: false\n  containers:\n  - name: demo\n    image: demo\n  securityContext:\n    allowPrivilegeEscalation: false\n    readOnlyRootFilesystem: true\n    runAsNonRoot: true\n",
        "level": "warning"
    }
]