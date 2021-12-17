import {SetHost} from "../../components/config"

export default function handler(req, res) {
    return new Promise((resolve) => {
        const {host} = JSON.parse(req.body)
        SetHost(host)
        res.status(200).send("true")
        resolve()
    })
}
