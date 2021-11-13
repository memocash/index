import styles from '../../styles/Home.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {PreInline} from "../../components/util/pre";

export default function Block() {
    const [blocks, setBlocks] = useState([])
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($newest: Boolean) {
        blocks(newest: $newest) {
            hash
            timestamp
            height
            txs {
                hash
            }
        }
    }
    `
    useEffect(() => {
        fetch("/api/graphql", {
            method: "POST",
            body: JSON.stringify({
                query: query
            }),
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
            console.log(data.data)
            setBlocks(data.data.blocks)
        }).catch(res => {
            setErrorMessage("error loading address")
            setLoading(true)
            console.log(res)
        })
    }, [])

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
                            <div key={block} className={column.container}>
                                <div className={column.width15}>Hash</div>
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
                </Loading>
            </div>
        </Page>
    )
}
