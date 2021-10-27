export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        res.status(200).json({message: "TX RESPONSE"})
        res.end()
        resolve()
    })
}
