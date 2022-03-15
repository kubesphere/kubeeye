import react from "react"
import "./index.scss"
import { RingGraph } from "./RingGraph/index"

export function ClusterStatus(prop) {
  const { data } = prop

  const scoreList = [
    { status: "", name: "健康检查项", count: data.total || 0 },
    { status: "normal", name: "正常项", count: data.passing || 0 },
    { status: "warning", name: "告警项", count: data.warning || 0 },
    { status: "dangerous", name: "危险项", count: data.dangerous || 0 },
  ]

  return (
    <div className="clusterStatus">
      <div className="title">集群健康状态</div>
      <div className="info">
        <div className="chart">
          <RingGraph data={data} />
        </div>
        <div className="list">
          {scoreList.map((item) => {
            return (
              <div className="tag">
                <div className="name">{item.name}</div>
                <div className="count">{item.count}</div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
