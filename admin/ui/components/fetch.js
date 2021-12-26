let host = ""

export function setHost(newHost) {
    host = newHost
}

export default async function getUrl(url, options) {
    if (options.headers === undefined) {
        options.headers = {}
    }
    options.headers.server = host
    return await fetch(url, options)
}
