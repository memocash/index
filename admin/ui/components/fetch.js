let host = ""

export function setHost(newHost) {
    host = newHost
}

export async function graphQL(query, variables) {
    return await getUrl("/api/graphql", {
        method: "POST",
        body: JSON.stringify({query, variables}),
    })
}

export default async function getUrl(url, options) {
    if (options.headers === undefined) {
        options.headers = {}
    }
    options.headers.server = host
    return await fetch(url, options)
}
