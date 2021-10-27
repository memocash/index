import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import Loading from "../../components/util/loading";

export default function Hash() {
    const [response, setResponse] = useState("")
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($hash: String!) {
        tx(hash: $hash) {
            hash
            raw
        }
    }
    `
    useEffect(() => {
        fetch("/api/graphql", {
            method: "POST",
            body: JSON.stringify({
                query: query,
                variables: {
                    hash: "hash-variable",
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
                            Code: {data.errors[i].extensions.code}
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
                return
            }
            setResponse(data.data.tx.hash)
            setLoading(false)
        }).catch(res => {
            res.json().then(data => {
                setErrorMessage(<>Code: {res.status}<br/>Message: {data.message}</>)
            })
        })
    }, [])
    return (
        <Page>
            <div>
                <h1 className={styles.title}>
                    Transaction
                </h1>
                <Loading loading={loading} error={errorMessage}>
                    <p>
                        Response: {response}
                    </p>
                </Loading>
            </div>
        </Page>
    )
}
