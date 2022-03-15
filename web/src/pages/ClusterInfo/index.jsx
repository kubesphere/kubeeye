import "./index.scss"

export function ClusterInfo(prop) {
  const { data } = prop
  const clusterInfo = [
    { name: "namespaces", count: data.namespacesCount },
    { name: "nodes", count: data.nodesCount },
    { name: "workloads", count: data.workloadsCount },
    { name: "Kubernetes version", count: data.version },
  ]
  return (
    <div className="clusterInfo">
      <div className="title">集群信息</div>
      <div className="list">
        {clusterInfo.map((item) => {
          return (
            <div className="tag">
              <div className="name">{item.name}</div>
              <div className="count">{item.count}</div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
