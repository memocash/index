import {GetHost} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        let {server} = req.headers
        if (!server || !server.length) {
            server = GetHost()
        }
        fetch(server + "/topic/list", {
            method: "POST",
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
