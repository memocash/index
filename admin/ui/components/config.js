const DevHost = "http://127.0.0.1:26770"
const LiveHost = "http://127.0.0.1:26771"
const LiveSvHost = "http://127.0.0.1:26772"
let Host = LiveHost


const SetHost = (host) => {
    Host = host
};

const GetHost = () => {
    return Host
};

export {
    SetHost,
    GetHost,
    DevHost,
    LiveHost,
    LiveSvHost,
}
