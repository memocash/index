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
    const router = useRouter()
    const query = `
    query ($newest: Boolean, $start: Uint32) {
        blocks(newest: $newest, start: $start) {
            hash
            timestamp
            height
            txs {
                hash
            }
        }
    }
    `
    let lastStart = null
    useEffect(() => {
        if (!router || !router.query || router.query.start === lastStart) {
            return
        }
        const {start} = router.query
        lastStart = start
        graphQL(query, {
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
            setBlocks(data.data.blocks)
            if (data.data.blocks.length > 0) {
                setNext(data.data.blocks[data.data.blocks.length - 1].height + 1)
                let previous = data.data.blocks[0].height - data.data.blocks.length;
                if (previous < 0) {
                    previous = 0
                }
                setPrev(previous)
            }
        }).catch(res => {
            setErrorMessage("error loading address")
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
                    <h3>Blocks </h3>
                    {blocks.map((block) => {
                        return (
                            <div key={block.hash} className={column.container}>
                                <div className={column.width15}>{block.height}</div>
                                <div className={column.width85}>
                                    <Link href={"/block/" + block.hash}>
                                        <a>
                                            <PreInline>{block.hash}</PreInline>
                                        </a>
                                    </Link>
                                </div>
                            </div>
                        )
                    })}
                    <div>
                        <Link href={{
                            pathname: "/block/list",
                            query: {
                                start: prev,
                            }
                        }} onClick={() => paginateClick(prev)}>
                            <a>Prev</a>
                        </Link>
                        <Link href={{
                            pathname: "/block/list",
                            query: {
                                start: next,
                            }
                        }} onClick={() => paginateClick(next)}>
                            <a>Next</a>
                        </Link>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
