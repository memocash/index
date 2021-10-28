import styles from '../../styles/Home.module.css'
import pre from '../../styles/pre.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import Loading from "../../components/util/loading";
import Link from "next/link";

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
            inputs {
                index
                prev_hash
                prev_index
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
            setErrorMessage("there was an error")
            return Promise.reject(res)
        }).then(data => {
            if (data.errors && data.errors.length > 0) {
                let messages = [];
                for (let i = 0; i < data.errors.length; i++) {
                    messages.push(
                        <p key={i}>
                            {data.errors[i].extensions ?
                                <>Code: {data.errors[i].extensions.code}</>
                                : null
                            }
                            <br/>
                            Message: {data.errors[i].message}
                        </p>
                    )
                }
                setErrorMessage(
                    <div>
                        {messages}
                    </div>
                )
                setLoading(true)
                return
            }
            setTx(data.data.tx)
            setLoading(false)
        }).catch(res => {
            res.json().then(data => {
                setErrorMessage(<>Code: {res.status}<br/>Message: {data.message}</>)
            })
        })
    }, [router])

    function preInline(children) {
        return (
            <pre className={[pre.pre, pre.inline].join(" ")}>
                {children}
            </pre>
        )
    }

    return (
        <Page>
            <div>
                <h1 className={styles.title}>
                    Transaction
                </h1>
                <Loading loading={loading} error={errorMessage}>
                    <div className={column.container}>
                        <div className={column.width15}>Tx hash</div>
                        <div className={column.width85}>{tx.hash}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Tx raw</div>
                        <div className={column.width85}>{preInline(<>{tx.raw}</>)}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width50}>
                            <h3>Inputs ({tx.inputs.length})</h3>
                            {tx.inputs.map((input) => {
                                return (
                                    <div key={input} className={column.container}>
                                        <div className={column.width15}>{input.index}</div>
                                        <div className={column.width85}>
                                            <Link href={"/tx/" + input.prev_hash}>
                                                <a>{preInline(<>{input.prev_hash}:{input.prev_index}</>)}</a>
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
                                            PkScript: <pre className={[pre.pre, pre.inline].join(" ")}>{output.script}</pre>
                                            {output.spends ? <>
                                                <h5>Spends ({output.spends.length})</h5>
                                                {output.spends.map((spend, index) => {
                                                    return (
                                                        <div key={index}>
                                                            <Link href={"/tx/" + spend.hash}>
                                                                <a>{preInline(<>{spend.hash}:{spend.index}</>)}</a>
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
