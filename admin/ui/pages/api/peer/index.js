export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {ip, port} = JSON.parse(req.body)
        fetch("http://127.0.0.1:26770/node/history", {
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
