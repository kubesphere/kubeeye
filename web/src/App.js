import {Loading, Table} from "@kube-design/components";
import {useEffect, useState} from 'react';
import "@kube-design/components/esm/styles/index.css";
import {getNamespaces} from "./services/namespace";

export function App() {
    const [namespaces, setNamespaces] = useState({});

    useEffect(() => {
        getNamespaces().then(function (nss) {
            setNamespaces(nss)
        })
    }, []);

    const columns = [{
        title: 'Name',
        dataIndex: 'metadata.name',
    }, {
        title: 'Create Time',
        dataIndex: 'metadata.creationTimestamp',
    }, {
        title: 'Status',
        dataIndex: 'status.phase',
    }]

    if (Object.keys(namespaces).length === 0) {
        return <Loading/>
    }
    return (
        <div>
            <Table rowKey="metadata.name" columns={columns} dataSource={namespaces}/>
        </div>
    );
}