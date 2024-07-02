import styles from '../../styles/Home.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {PreInline} from "../../components/util/pre";
import {useRouter} from "next/router";
import {graphQL} from "../../components/fetch";

export default function Block() {
    const [blocks, setBlocks] = useState([])
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const [next, setNext] = useState(4)
    const [prev, setPrev] = useState(0)
    const [lastStart, setLastStart] = useState(0)
    const [lastNewest, setLastNewest] = useState(true)
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
        if (!router || !router.query ||
            (router.query.start === lastStart.toString() && router.query.newest === lastNewest.toString())) {
            return
        }
        let {start, newest} = router.query
        if (!start) {
            start = 0
        }
        newest = newest !== "false"
        setLastStart(start)
        setLastNewest(newest)
        graphQL(query, {
            start: start,
            newest: newest,
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
                if (newest) {
                    setNext(data.data.blocks[data.data.blocks.length - 1].height)
                    let previous = data.data.blocks[0].height + data.data.blocks.length + 1;
                    if (previous < 0) {
                        previous = 0
                    }
                    setPrev(previous)
                } else {
                    setNext(data.data.blocks[data.data.blocks.length - 1].height + 1)
                    let previous = data.data.blocks[0].height - data.data.blocks.length;
                    if (previous < 0) {
                        previous = 0
                    }
                    setPrev(previous)
                }
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
                <p>
                    <Link href={{
                        pathname: "/block/list",
                        query: {
                            newest: true,
                        }
                    }}>
                        <span className={lastNewest ? styles.underline : null}>Newest</span>
                    </Link>
                    &nbsp;&middot;&nbsp;
                    <Link href={{
                        pathname: "/block/list",
                        query: {
                            newest: false,
                        }
                    }}>
                        <span className={lastNewest ? null : styles.underline}>Oldest</span>
                    </Link>
                </p>
                <Loading loading={loading} error={errorMessage}>
                    <h3>Blocks </h3>
                    {blocks.map((block) => {
                        return (
                            <div key={block.hash} className={column.container}>
                                <div className={column.width15}>{block.height}</div>
                                <div className={column.width40}>
                                    <Link href={"/block/" + block.hash}>
                                        <PreInline>{block.hash}</PreInline>
                                    </Link>
                                </div>
                                <div className={[column.width15].join(" ")}>
                                    {block.timestamp}
                                </div>
                                <div className={[column.width15, column.right].join(" ")}>
                                    {block.size ? block.size.toLocaleString() : 0} bytes
                                </div>
                                <div className={[column.width15, column.right].join(" ")}>
                                    {block.tx_count ? block.tx_count.toLocaleString() : 0} txs
                                </div>
                            </div>
                        )
                    })}
                    <div>
                        <Link href={{
                            pathname: "/block/list",
                            query: {
                                newest: lastNewest,
                                start: prev,
                            }
                        }}>
                            Prev
                        </Link>
                        &nbsp;&middot;&nbsp;
                        <Link href={{
                            pathname: "/block/list",
                            query: {
                                newest: lastNewest,
                                start: next,
                            }
                        }}>
                            Next
                        </Link>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
