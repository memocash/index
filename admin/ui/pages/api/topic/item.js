import {GetHost} from "../../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {topic, shard, uid} = JSON.parse(req.body)
        const shardInt = parseInt(shard)
        fetch(GetHost() + "/topic/item", {
            method: "POST",
            body: JSON.stringify({
                Topic: topic,
                Shard: shardInt,
                Uid: uid,
            }),
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
