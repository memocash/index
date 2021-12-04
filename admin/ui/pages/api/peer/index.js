import {Host} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {ip, port} = JSON.parse(req.body)
        fetch(Host + "/node/history", {
            method: "POST",
            body: JSON.stringify({
                Ip: ip,
                Port: port,
            })
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
