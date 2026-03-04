import Page from "../../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import {GetErrorMessage} from "../../../components/util/loading";
import {graphQL} from "../../../components/fetch";
import BlockDetail from "../../../components/block-detail";

export default function BlockHeight() {
    const router = useRouter()
    const [block, setBlock] = useState({
        hash: "",
        height: 0,
        timestamp: "",
        txs: [],
    })
    const [lastHeight, setLastHeight] = useState(undefined)
    const [lastOffset, setLastOffset] = useState(0)
    const [offset, setOffset] = useState(0)
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($height: Int!, $start: Uint32) {
        block_by_height(height: $height) {
            hash
            height
            timestamp
            raw
            size
            tx_count
            txs(start: $start) {
                index
                tx {
                    hash
                }
            }
        }
    }
    `
    useEffect(() => {
        if (!router || !router.query || !router.query.height) {
            return
        }
        const {height, start} = router.query
        const heightInt = parseInt(height)
        if (heightInt === lastHeight && start === lastOffset) {
            return
        }
        setLastHeight(heightInt)
        if (start) {
            setLastOffset(parseInt(start))
        } else {
            setLastOffset(0)
        }
        setLoading(true)
        setErrorMessage("")
        graphQL(query, {
            height: heightInt,
            start: start ? start : undefined,
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
            setBlock(data.data.block_by_height)
            if (data.data.block_by_height.txs && data.data.block_by_height.txs.length > 0) {
                setOffset(data.data.block_by_height.txs[data.data.block_by_height.txs.length - 1].index + 1)
            }
        }).catch(res => {
            setErrorMessage("error loading block")
            setLoading(true)
            console.log(res)
        })
    }, [router])
    return (
        <Page>
            <BlockDetail block={block} lastOffset={lastOffset} offset={offset}
                         hashLink={true} txPaginationPath={"/block/height/" + block.height}
                         loading={loading} errorMessage={errorMessage}/>
        </Page>
    )
}
