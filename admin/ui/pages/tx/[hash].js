import styles from '../../styles/Home.module.css'
import pre from '../../styles/pre.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import {Loading, GetErrorMessage} from "../../components/util/loading";
import Link from "next/link";
import {PreInline} from "../../components/util/pre";

export default function Hash() {
    const router = useRouter()
    const [tx, setTx] = useState({
        inputs: [],
        outputs: [],
    })
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($hash: String!) {
        tx(hash: $hash) {
            hash
            raw
            suspect {
                hash
            }
            lost {
                hash
            }
            inputs {
                index
                prev_hash
                prev_index
                output {
                    spends {
                        hash
                        index
                    }
                }
            }
            outputs {
                index
                amount
                script
                spends {
                    hash
                    index
                }
            }
        }
    }
    `
    let lastHash = undefined
    useEffect(() => {
        if (!router || !router.query || router.query.hash === lastHash) {
            return
        }
        const {hash} = router.query
        lastHash = hash
        fetch("/api/graphql", {
            method: "POST",
            body: JSON.stringify({
                query: query,
                variables: {
                    hash: hash,
                }
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
            setTx(data.data.tx)
        }).catch(res => {
            setErrorMessage("error loading tx")
            setLoading(true)
            console.log(res)
        })
    }, [router])

    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    Transaction
                </h2>
                <Loading loading={loading} error={errorMessage}>
                    <div className={column.container}>
                        <div className={column.width15}>Tx hash</div>
                        <div className={column.width85}>{tx.hash}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Tx raw</div>
                        <div className={column.width85}><PreInline>{tx.raw}</PreInline></div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Lost/suspect</div>
                        <div className={column.width85}>
                            {tx.lost ?
                                <div className={[column.red, column.bold].join(" ")}>
                                    LOST
                                </div>
                                : (tx.suspect ?
                                    <div className={[column.orange, column.bold].join(" ")}>
                                        SUSPECT
                                    </div>
                                    : "OK!")}
                        </div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width50}>
                            <h3>Inputs ({tx.inputs.length})</h3>
                            {tx.inputs.map((input) => {
                                return (
                                    <div key={input} className={column.container}>
                                        <div className={column.width15}>{input.index}</div>
                                        <div className={column.width85}>
                                            Amount: {input.output.amount}
                                            <br/>
                                            {input.output.spends && input.output.spends.length >= 2 ?
                                                <div className={[column.red, column.bold].join(" ")}>
                                                    DOUBLE SPEND
                                                </div>
                                                : null
                                            }
                                            <Link href={"/tx/" + input.prev_hash}>
                                                <a><PreInline>{input.prev_hash}:{input.prev_index}</PreInline></a>
                                            </Link>
                                        </div>
                                    </div>
                                )
                            })}
                        </div>
                        <div className={column.width50}>
                            <h3>Outputs ({tx.outputs.length})</h3>
                            {tx.outputs.map((output, index) => {
                                return (
                                    <div key={index} className={column.container}>
                                        <div className={column.width15}>
                                            {output.index}
                                        </div>
                                        <div className={column.width85}>
                                            Amount: {output.amount}
                                            <br/>
                                            PkScript: <pre
                                            className={[pre.pre, pre.inline].join(" ")}>{output.script}</pre>
                                            {output.spends ? <>
                                                {output.spends.length >= 2 ?
                                                    <div className={[column.red, column.bold].join(" ")}>
                                                        DOUBLE SPEND
                                                    </div>
                                                    : null
                                                }
                                                <h5 className={column.noMargin}>Spends
                                                    ({output.spends.length})</h5>
                                                {output.spends.map((spend, index) => {
                                                    return (
                                                        <div key={index}>
                                                            <Link href={"/tx/" + spend.hash}>
                                                                <a>
                                                                    <PreInline>{spend.hash}:{spend.index}</PreInline>
                                                                </a>
                                                            </Link>
                                                        </div>
                                                    )
                                                })}
                                            </> : null}
                                        </div>
                                    </div>
                                )
                            })}
                        </div>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
