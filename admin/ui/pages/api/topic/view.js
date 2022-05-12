import {GetHost} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {topic} = JSON.parse(req.body)
        fetch(GetHost() + "/topic/view", {
            method: "POST",
            body: JSON.stringify({
                Topic: topic
            }),
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
