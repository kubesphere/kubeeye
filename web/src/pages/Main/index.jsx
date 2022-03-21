import React, { useEffect } from "react"
import { useState } from "react"
import { Select } from "@kube-design/components"
import { isEmpty } from "lodash"
import Empty from "../components/Empty"
import classnames from "classnames"
import './index.scss';
import { ClusterItemCard } from "./ClusterItemCard"
import { ProjectCheckCard } from './ProjectCheckCard'
export function Main(prop) {
  const { data } = prop
  const [currentIndex, setCurrentIndex] = useState(1)
  const [proData,setProData] = useState([])
  const [curProData, setCurProData] = useState([]);

  useEffect(()=>{
    const projectData = data.filter((item) => {
        return item.namespace !== ""
    })
    setProData(projectData);
    setCurProData(projectData);
  },[])

    const clusterData = data.filter((item) => {
        return item.namespace === ""
    })

    let programOptions = []
     const projectData = data.filter((item) => {
        return item.namespace !== ""
    })
    projectData.map(item => {
      let obj = {
          label: item.namespace,
          value: item.namespace
      }
      programOptions.push(obj)
    })

    const searchByName = (name) => {
       const pro = proData.filter(item =>{
           return item.namespace === name
       })
       setCurProData(pro)
    }

    const renderClusterCheck = () => {
    const checkData = clusterData[0]?.resultInfos;
    if (isEmpty(checkData)) {
      return <Empty />
    }
    return  (
       <div>
           {checkData.map((cluster,index) =>{
               return (
                   <ClusterItemCard
                        key={index}
                        detail={cluster}
                        expand={index === 0}
                    />
               )
           })}
       </div>
    )
    }

    const renderProjectCheck = () => {
        if (isEmpty(proData)) {
            return <Empty />
        }

        const handleKeyChange = (value) => {
            searchByName(value)
        }
        return <div className="programCheck">
            <div className="allProject">
            <div className="project">
                <Select
                    name="project"
                    defaultValue=""
                    placeholder="全部项目"
                    onChange={handleKeyChange}
                    options={programOptions}
                ></Select>
             </div>
            </div>
            <div className="projectContent">
                {curProData[0].resultInfos.map((cluster,index) => {
                    return (
                        <ProjectCheckCard
                            key={index}
                            detail={cluster}
                            expand={index === 0}
                        />
                    )
                })}
            </div>
        </div>
    }

    const handleChange = (id) => {
      setCurrentIndex(id);
    }

    let isShowCluster = currentIndex === 1 ? 'block': 'none';
    let isShowProject = currentIndex === 2 ? 'block' : 'none';

    const tabs = [
       { tabName: "集群检查", id: 1 },
       { tabName: "项目检查", id: 2 },
   ]

    return (
        <div className="main">
            <div className="selectCheck">
                {tabs.map(item=>{
                    return (
                        <div className={classnames('clusterName', {     
                        'active': item.id === currentIndex
                        })}  onClick={handleChange.bind(this, item.id)}>{item.tabName}</div>
                    )
                })}
            </div>
            <div className="contentRender">
                <div style={{ "display": isShowCluster}}>
                    {renderClusterCheck()}
                </div>
                <div style={{ "display": isShowProject}}>
                    {renderProjectCheck()}
                </div>
            </div>
        </div>
    )
}
