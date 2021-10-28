import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import Loading from "../../components/util/loading";
import Link from "next/link";

export default function Hash() {
    const router = useRouter()
    const [tx, setTx] = useState({
        inputs: []
    })
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($hash: String!) {
        tx(hash: $hash) {
            hash
            raw
            inputs {
                prev_hash
                prev_index
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
                                : ""
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
    const preStyle = {
        wordWrap: "break-word",
        overflowWrap: "anywhere",
        whiteSpace: "pre-wrap",
        padding: 0,
        margin: 0,
    }
    return (
        <Page>
            <div>
                <h1 className={styles.title}>
                    Transaction
                </h1>
                <Loading loading={loading} error={errorMessage}>
                    <div>
                        Tx hash: {tx.hash}
                    </div>
                    <div>
                        Tx raw: <pre style={preStyle}>{tx.raw}</pre>
                    </div>
                    <div>
                        {tx.inputs.map((input) => {
                            return (
                                <p key={input}>
                                    <Link href={"/tx/"+input.prev_hash}>
                                        <a>{input.prev_hash}:{input.prev_index}</a>
                                    </Link>
                                </p>
                            )
                        })}
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
