import axios from 'axios';

export function getNamespaces() {
    return axios.get(`/api/v1/namespaces`)
        .then(res => {
            console.log('res-namespace', res)
            return res.data.items
        })
}

export function getData() {
    return axios.get(`/apis/kubeeye.kubesphere.io/v1alpha1/namespaces/default/clusterinsights/clusterinsight-sample`).then(
        res => {
            return res.data
        }
    )
}