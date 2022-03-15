import React from "react"
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts"

const COLORS = ["#55BC8A", "#CA2621", "#F5A623"]

export function RingGraph(prop) {
  const { data } = prop

  const graphData = [
    { name: "正常项", value: data.passing || 0 },
    { name: "告警项", value: data.warning || 0 },
    { name: "危险项", value: data.dangerous || 0 },
  ]
  return (
    <ResponsiveContainer width="100%" height="100%">
      <PieChart width={400} height={267}>
        <Pie
          data={graphData}
          cx={140}
          cy={140}
          innerRadius={60}
          outerRadius={80}
          fill="#8884d8"
          paddingAngle={5}
          dataKey="value"
        >
          {graphData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
      </PieChart>
    </ResponsiveContainer>
  )
}
