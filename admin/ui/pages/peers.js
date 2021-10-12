import Page from "../components/page";
import {useEffect, useState} from "react";

function Peers() {
    const [loading, setLoading] = useState(true)
    const [peers, setPeers] = useState([])
    const [errorMessage, setErrorMessage] = useState("")
    useEffect(() => {
        fetch("/api/peers")
            .then(res => {
                if (res.ok) {
                    return res.json()
                }
                return Promise.reject(res)
            })
            .then(data => {
                var peerCount = data.Peers.length;
                const MaxPeers = 10;
                if (peerCount > MaxPeers) {
                    peerCount = MaxPeers;
                }
                setPeers(data.Peers.splice(0, peerCount))
                setLoading(false)
            })
            .catch(res => {
                res.text().then(msg => {
                    setErrorMessage(<>Code: {res.status}<br/>Message: {msg}</>)
                })
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Peers Page
                </h1>
                {loading ?
                    <>{!!errorMessage ?
                        <>Error: {errorMessage}</>
                        :
                        <>Loading...</>
                    }</>
                    :
                    <div>
                        {peers.map(peer => (
                            <p>{peer.Ip}:{peer.Port}</p>
                        ))}
                    </div>
                }
            </div>
        </Page>
    )
}

export default Peers
