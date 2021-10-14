import Page from "../../components/page";
import {useEffect, useState} from "react";

function Report() {
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const [peersAttempted, setPeersAttempted] = useState(0)
    const [peersConnected, setPeersConnected] = useState(0)
    const [peersFailed, setPeersFailed] = useState(0)
    const [totalAttempts, setTotalAttempts] = useState(0)
    const [totalPeers, setTotalPeers] = useState(0)
    useEffect(() => {
        fetch("/api/peer/report").then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            setPeersAttempted(data.PeersAttempted)
            setPeersConnected(data.PeersConnected)
            setPeersFailed(data.PeersFailed)
            setTotalAttempts(data.TotalAttempts)
            setTotalPeers(data.TotalPeers)
            setLoading(false)
        }).catch(res => {
            res.text().then(msg => {
                setErrorMessage(<>Code: {res.status}<br/>Message: {msg}</>)
            })
        })
    }, [])

    return (
        <Page>
            <div>
                <h1>
                    Peer Report
                </h1>
                {loading ?
                    <>{!!errorMessage ?
                        <>Error: {errorMessage}</>
                        :
                        <>Loading...</>
                    }</>
                    :
                    <ul>
                        <li>Peers Attempted: {peersAttempted}</li>
                        <li>Peers Connected: {peersConnected}</li>
                        <li>Peers Failed: {peersFailed}</li>
                        <li>Total Attempts: {totalAttempts}</li>
                        <li>Total Peers: {totalPeers}</li>
                    </ul>
                }
            </div>
        </Page>
    )
}

export default Report
