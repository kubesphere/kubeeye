import axios from 'axios';

export function getNamespaces() {
    return axios.get(`/api/v1/namespaces`)
        .then(res => {
            return res.data.items
        })
}