export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        fetch("http://127.0.0.1:26770/node/peer_report").then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
