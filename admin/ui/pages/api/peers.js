import {GetHost} from "../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {filter} = JSON.parse(req.body)
        fetch(GetHost() + "/node/peers", {
            method: "POST",
            body: JSON.stringify({
                Filter: filter
            })
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
