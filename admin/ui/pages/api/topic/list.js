import {GetHost} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        fetch(GetHost() + "/topic/list", {
            method: "POST",
        }).then(res => res.json()).then(data => {
            console.log(data)
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
