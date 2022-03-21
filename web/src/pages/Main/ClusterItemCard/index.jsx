import Card from "../Card"
import classnames from "classnames"
import "./index.scss"
import { Icon } from "@kube-design/components"
import { useState } from "react"

export function ClusterItemCard(prop) {
  const [isExpand, setIsExpand] = useState(false)
  const { detail } = prop;
  const renderTitle = () => {
    return (
      <div className="mainContent">
        <div className="arrow">
            <Icon name="caret-right" size={12} type="dark" />
        </div>
        <div className="clusterName">
            <Icon name="cluster" size={40}/>
        </div>
        <div className="tableContent">
            <div className="text">
                <div className="name">name</div>
                <div className="value">{detail?.resultInfos?.name || '-'}</div>
            </div>
            <div className="text">
                <div className="name">resourceType</div>
                <div className="value">{detail.resourceType}</div>
            </div>
        </div>
      </div>
    )
  }

  handleExpand = () => {
      setIsExpand(!isExpand)
  }

  const renderContent = () => {
      const {resourceInfos} = detail;

      return (
          <div className="liContent"
             onClick={e=> e.stopPropagation()}
          >
            {resourceInfos && resourceInfos.items && resourceInfos.items.map(item => {
                return (
                    <div key={item.message} className="checklist"> 
                        <div className="message">{item.message}</div>
                        <div className="level">{item.level}</div>
                    </div>
                )
              })}
          </div>
      )
  }

  return (
    <div>
      <Card
        className={classnames("clusterContent", {
          "expanded": isExpand,
        })}
        title={renderTitle()}
        empty={'空数据'}
        onClick={handleExpand}
      >
          {renderContent()}
      </Card>
    </div>
  )
}
