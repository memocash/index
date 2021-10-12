import Page from "../components/page";
import {useEffect, useState} from "react";
import Pagination from "../components/util/pagination";

function Peers() {
    const [loading, setLoading] = useState(true)
    const [allPeers, setAllPeers] = useState([])
    const [peers, setPeers] = useState([])
    const [errorMessage, setErrorMessage] = useState("")
    const [totalPeers, setTotalPeers] = useState(0)
    useEffect(() => {
        fetch("/api/peers")
            .then(res => {
                if (res.ok) {
                    return res.json()
                }
                return Promise.reject(res)
            })
            .then(data => {
                setAllPeers(data.Peers);
                setTotalPeers(data.Peers.length);
                setLoading(false)
            })
            .catch(res => {
                res.text().then(msg => {
                    setErrorMessage(<>Code: {res.status}<br/>Message: {msg}</>)
                })
            })
    }, [])

    const onPageChanged = (data) => {
        const {currentPage, totalPages, pageLimit} = data;
        const offset = (currentPage - 1) * pageLimit;
        setPeers(allPeers.slice(offset, offset + pageLimit))
    }

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
                        <div>
                            {peers.map(peer => (
                                <p>{peer.Ip}:{peer.Port}</p>
                            ))}
                        </div>
                        <Pagination totalRecords={totalPeers} pageLimit={10} pageNeighbours={0} onPageChanged={onPageChanged}/>
                    </div>
                }
            </div>
        </Page>
    )
}

export default Peers
