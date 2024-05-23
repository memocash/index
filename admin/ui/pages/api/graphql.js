import {GetHostGraphQL} from "../../components/config"

export default function handler(req, res) {
    return new Promise((resolve, reject) => {
        const {query, variables} = JSON.parse(req.body)
        let {server} = req.headers
        if (!server || !server.length) {
            server = GetHostGraphQL()
        }
        fetch(server + "/graphql", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                query: query,
                variables: variables,
            })
        }).then(res => res.json()).then(data => {
            res.status(200).json(data)
            resolve()
        }).catch(error => {
            reject(error)
        })
    })
}
