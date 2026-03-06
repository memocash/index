import styles from '../../styles/Home.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {useRouter} from "next/router";
import {graphQL} from "../../components/fetch";

export default function Block() {
    const [blocks, setBlocks] = useState([])
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const [older, setOlder] = useState(0)
    const [newer, setNewer] = useState(0)
    const [hasNewer, setHasNewer] = useState(false)
    const [hasOlder, setHasOlder] = useState(false)
    const [lastStart, setLastStart] = useState(null)
    const router = useRouter()
    const query = `
    query ($newest: Boolean, $start: Uint32) {
        blocks(newest: $newest, start: $start) {
            hash
            timestamp
            height
            size
            tx_count
        }
    }
    `
    useEffect(() => {
        if (!router || !router.query) {
            return
        }
        let {start} = router.query
        const startKey = start || ""
        if (startKey === lastStart) {
            return
        }
        const startInt = parseInt(start) || 0
        setLastStart(startKey)
        graphQL(query, {
            start: startInt,
            newest: true,
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
            setBlocks(data.data.blocks)
            if (data.data.blocks.length > 0) {
                const firstHeight = data.data.blocks[0].height
                const lastHeight = data.data.blocks[data.data.blocks.length - 1].height
                const pageSize = data.data.blocks.length
                setOlder(lastHeight)
                setNewer(firstHeight + pageSize + 1)
                setHasOlder(lastHeight >= pageSize)
                setHasNewer(startInt > 0 && firstHeight + 1 >= startInt)
            }
        }).catch(res => {
            setErrorMessage("error loading blocks")
            setLoading(true)
            console.log(res)
        })
    }, [router])
    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    Blocks ({blocks.length})
                </h2>
                <Loading loading={loading} error={errorMessage}>
                    <div className={[column.container, column.singleLine, column.bold].join(" ")}>
                        <div className={column.fixedHeight}>Height</div>
                        <div className={column.flexFill}>Hash</div>
                        <div className={[column.flexFit, column.truncate].join(" ")}>Timestamp</div>
                        <div className={[column.fixedSize, column.right].join(" ")}>Size (bytes)</div>
                        <div className={[column.fixedTxCount, column.right].join(" ")}>Txs</div>
                    </div>
                    {blocks.map((block) => {
                        return (
                            <div key={block.hash} className={[column.container, column.singleLine].join(" ")}>
                                <div className={column.fixedHeight}>
                                    <Link href={"/block/height/" + block.height}>
                                        {block.height}
                                    </Link>
                                </div>
                                <div className={[column.flexFill, column.truncate, column.monospace].join(" ")}>
                                    <Link href={"/block/" + block.hash}>
                                        {block.hash}
                                    </Link>
                                </div>
                                <div className={[column.flexFit, column.truncate].join(" ")}>
                                    {block.timestamp.replace(/[-+]\d{2}:\d{2}$/, "")}
                                </div>
                                <div className={[column.fixedSize, column.right].join(" ")}>
                                    {block.size ? block.size.toLocaleString() : 0}
                                </div>
                                <div className={[column.fixedTxCount, column.right].join(" ")}>
                                    {block.tx_count ? block.tx_count.toLocaleString() : 0}
                                </div>
                            </div>
                        )
                    })}
                    <div className={column.navButtons}>
                        {hasNewer ?
                            <Link href={{pathname: "/block/list"}}>
                                <span className={column.navButton}>Newest</span>
                            </Link>
                            : <span className={column.navButtonDisabled}>Newest</span>
                        }
                        {" "}
                        {hasNewer ?
                            <Link href={{pathname: "/block/list", query: {start: newer}}}>
                                <span className={column.navButton}>&laquo; Newer</span>
                            </Link>
                            : <span className={column.navButtonDisabled}>&laquo; Newer</span>
                        }
                        {" "}
                        {hasOlder ?
                            <Link href={{pathname: "/block/list", query: {start: older}}}>
                                <span className={column.navButton}>Older &raquo;</span>
                            </Link>
                            : <span className={column.navButtonDisabled}>Older &raquo;</span>
                        }
                        {" "}
                        {hasOlder ?
                            <Link href={{pathname: "/block/list", query: {start: 26}}}>
                                <span className={column.navButton}>Oldest</span>
                            </Link>
                            : <span className={column.navButtonDisabled}>Oldest</span>
                        }
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
