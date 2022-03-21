import {
    Loading,
    Table,
    Button,
    Select
} from "@kube-design/components";
import {
    useEffect,
    useState
} from 'react';
import "@kube-design/components/esm/styles/index.css";
import {
    getData,
    getNamespaces
} from "./services/namespace";
import './App.scss'
import Button from "@kube-design/components/lib/components/Button";
import { ClusterStatus } from "./pages/ClusterStatus";
import { ClusterInfo } from "./pages/ClusterInfo";

import{ Main } from './pages/Main';
export function App() {
    const [namespaces, setNamespaces] = useState({});
    const [proData,setProData] = useState({})

    useEffect(() => {
        // getNamespaces().then(function (nss) {
        //     setNamespaces(nss)
        // })
        getData().then(function(data){
            setProData(data.status);
            console.log('data',data)
        })
    }, []);
    console.log('proData', proData)
    const { clusterInfo={}, scoreInfo={}, auditResults=[]} = proData;
    console.log('auditResults', auditResults)
    return (
        <div className="container">
            <div className="header">
                <div className="logo">
                    <img src="/assets/kubeEye-logo.svg" alt=''></img>
                </div>
                <Button className="lauguage">
                    <img src="/assets/kubeeye-logo.png" alt=''>
                    </img>
                </Button>
            </div>
            <div className="content">
                <div className="clusterTop">
                    <ClusterStatus data={scoreInfo}/>
                    <ClusterInfo data={clusterInfo}/>
                </div>
                <Main data={auditResults}/>
            </div>
        </div>
    );
}