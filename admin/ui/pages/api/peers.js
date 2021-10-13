export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        fetch("http://127.0.0.1:26770/node/peers", {
            method: "POST",
            body: JSON.stringify({
                Page: 0
            })
        })
            .then(res => res.json())
            .then(data => {
                res.status(200).json(data)
                resolve()
            })
            .catch(error => {
                reject(error)
            })
    })
}
