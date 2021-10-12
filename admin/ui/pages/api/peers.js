export default function handler(req, res) {
    //res.status(200).json({name: 'Hello'})
    fetch("http://127.0.0.1:26770/node/peers", {
        body: JSON.stringify({
            Page: 0
        })
    })
        .then(res => res.json())
        .then(data => res.status(200).json(data))
}
