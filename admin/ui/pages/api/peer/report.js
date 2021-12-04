import {Host} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        fetch(Host + "/node/peer_report").then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
