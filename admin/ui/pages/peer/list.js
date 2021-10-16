import Page from "../../components/page";
import Pagination from "../../components/util/pagination";
import {useEffect, useRef, useState} from "react";
import Link from 'next/link';
import styles from '../../styles/list.module.css';
import dropdownStyles from '../../styles/dropdown.module.css';

function List() {
    const [loading, setLoading] = useState(true)
    const [allPeers, setAllPeers] = useState([])
    const [peers, setPeers] = useState([])
    const [errorMessage, setErrorMessage] = useState("")
    const [totalPeers, setTotalPeers] = useState(0)
    const [filterValue, setFilterValue] = useState("all")
    const PageLimit = 10
    const inputPagination = useRef(null)

    useEffect(() => {
        const urlSearchParams = new URLSearchParams(window.location.search)
        const params = Object.fromEntries(urlSearchParams.entries())
        if (params.filter) {
            setFilterValue(params.filter)
        }
        if (params.page) {
            inputPagination.gotoPage(params.page)
        }
    }, [])

    useEffect(() => {
        fetch("/api/peers", {
            method: "POST",
            body: JSON.stringify({
                filter: filterValue,
            })
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            setAllPeers(data.Peers)
            setTotalPeers(data.Peers.length)
            setPagePeers(0, data.Peers)
            setLoading(false)
        }).catch(res => {
            res.text().then(msg => {
                setErrorMessage(<>Code: {res.status}<br/>Message: {msg}</>)
            })
        })
    }, [filterValue])

    const onPageChanged = (data) => {
        const {currentPage} = data
        let searchParams = new URLSearchParams(window.location.search);
        searchParams.set("page", currentPage)
        window.history.push({
            pathname: window.location.pathname,
            search: searchParams.toString(),
        })
        setPagePeers((currentPage - 1) * PageLimit, allPeers)
    }

    const setPagePeers = (offset, tempAllPeers) => {
        setPeers(tempAllPeers.slice(offset, offset + PageLimit))
    }

    return (
        <Page>
            <div>
                <h1>
                    Peers Page
                </h1>
                <div>
                    <select className={dropdownStyles.select} onChange={e => setFilterValue(e.target.value)}
                            value={filterValue}>
                        <option value={"all"}>All</option>
                        <option value={"attempted"}>Attempted</option>
                        <option value={"successes"}>Successes</option>
                    </select>
                </div>
                {loading ?
                    <>{!!errorMessage ?
                        <>Error: {errorMessage}</>
                        :
                        <>Loading...</>
                    }</>
                    :
                    <div>
                        <ul className={styles.list}>
                            {peers.map((peer, key) => (
                                <li key={key}>
                                    <Link href={{
                                        pathname: "/peer/view",
                                        query: {
                                            ip: peer.Ip,
                                            port: peer.Port
                                        }
                                    }}>
                                        <a>{peer.Ip}:{peer.Port} - {peer.Time} - {peer.Status}</a>
                                    </Link>
                                </li>
                            ))}
                        </ul>
                        <Pagination ref={inputPagination} totalRecords={totalPeers} pageLimit={PageLimit}
                                    pageNeighbours={1}
                                    onPageChanged={onPageChanged}/>
                    </div>
                }
            </div>
        </Page>
    )
}

export default List
