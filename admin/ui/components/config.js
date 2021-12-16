const DevHost = "http://127.0.0.1:26770"
const LiveHost = "http://127.0.0.1:26771"
let Host = LiveHost


const SetHost = (host) => {
    Host = host
};

export {
    SetHost,
    DevHost,
    LiveHost,
    Host,
}
