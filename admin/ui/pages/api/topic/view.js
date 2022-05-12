import {GetHost} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {topic, start, shard} = JSON.parse(req.body)
        const shardInt = parseInt(shard)
        fetch(GetHost() + "/topic/view", {
            method: "POST",
            body: JSON.stringify({
                Topic: topic,
                Start: start,
                Shard: shardInt,
            }),
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
