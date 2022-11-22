import styles from '../../styles/Home.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {graphQL} from "../../components/fetch";

export default function LockHash() {
    const router = useRouter()
    const [block, setBlock] = useState({
        hash: "",
        height: 0,
        timestamp: "",
        txs: [],
    })
    const [lastStart, setLastStart] = useState(undefined)
    const [lastTx, setLastTx] = useState("")
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($hash: String!, $start: Uint32) {
        block(hash: $hash) {
            hash
            height
            timestamp
            raw
            size
            tx_count
            txs(start: $start) {
                index
                hash
            }
        }
    }
    `
    let lastBlockHash = undefined
    useEffect(() => {
        if (!router || !router.query || (router.query.hash === lastBlockHash && router.query.start === lastStart)) {
            return
        }
        const {hash, start} = router.query
        lastBlockHash = hash
        setLastStart(start)
        graphQL(query, {
            hash: hash,
            start: start,
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            if (data.errors && data.errors.length > 0) {
                setErrorMessage(GetErrorMessage(data.errors))
                setLoading(true)
                return
            }
            setLoading(false)
            setBlock(data.data.block)
            if (data.data.block.txs.length > 0) {
                setLastTx(data.data.block.txs[data.data.block.txs.length - 1].index)
            }
        }).catch(res => {
            setErrorMessage("error loading block")
            setLoading(true)
            console.log(res)
        })
    }, [router])
    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    Block
                </h2>
                <Loading loading={loading} error={errorMessage}>
                    <div className={column.container}>
                        <div className={column.width15}>Hash</div>
                        <div className={column.width85}>{block.hash}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Timestamp</div>
                        <div className={column.width85}>{block.timestamp}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Height</div>
                        <div className={column.width85}>{block.height.toLocaleString()}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Raw</div>
                        <div className={column.width85}>
                            <pre className={column.pre}>{block.raw}</pre>
                        </div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Size</div>
                        <div className={column.width85}>{block.size ? block.size.toLocaleString() : 0} bytes</div>
                    </div>
                    <div className={column.container}>
                        <div>
                            <h3>Txs ({block.txs.length} of {block.tx_count ? block.tx_count.toLocaleString() : 0})</h3>
                            {block.txs.map((tx) => {
                                return (
                                    <div key={tx.hash} className={column.container}>
                                        <div>
                                            <Link href={"/tx/" + tx.hash}>
                                                <a>{tx.hash}</a>
                                            </Link>
                                        </div>
                                    </div>
                                )
                            })}
                        </div>
                    </div>
                    <div>
                        <Link href={{pathname: "/block/" + block.hash}}>
                            <a>First</a>
                        </Link>
                        &nbsp;&middot;&nbsp;
                        <Link href={{
                            pathname: "/block/" + block.hash,
                            query: {
                                start: lastTx,
                            }
                        }}>
                            <a>Next</a>
                        </Link>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
