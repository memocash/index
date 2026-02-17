import {DevHost, LiveHost, LiveSvHost} from "./config";

function getInitialHost() {
    if (typeof window === 'undefined') return DevHost
    const val = localStorage.getItem("server-select")
    switch (val) {
        case "live":
            return LiveHost
        case "live-sv":
            return LiveSvHost
        default:
            return LiveHost
    }
}

let host = getInitialHost()

export function setHost(newHost) {
    host = newHost
}

export async function graphQL(query, variables) {
    return await getUrl("/api/graphql", {
        method: "POST",
        body: JSON.stringify({query, variables}),
    }, true)
}

export async function getUrl(url, options, graph=false) {
    if (!options) {
        options = {}
    }
    if (options.headers === undefined) {
        options.headers = {}
    }
    options.headers.server = graph ? host : "http://127.0.0.1:26768"
    return await fetch(url, options)
}
