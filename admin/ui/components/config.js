const DevHost = "http://127.0.0.1:26770"
const LiveHost = "https://graph.cash"
const LiveSvHost = "http://127.0.0.1:26772"
let Host = DevHost


const SetHost = (host) => {
    Host = host
};

const GetHost = () => {
    return "http://127.0.0.1:26768"
};

const GetHostGraphQL = () => {
    return Host
};

export {
    SetHost,
    GetHost,
    GetHostGraphQL,
    DevHost,
    LiveHost,
    LiveSvHost,
}
