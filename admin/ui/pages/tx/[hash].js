import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import Loading from "../../components/util/loading";

export default function Hash() {
    const [response, setResponse] = useState("")
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const router = useRouter()
    /*useEffect(() => {
        fetch("/api/tx", {
            method: "POST",
            body: JSON.stringify({
                hash: router.query.hash,
            }),
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            setErrorMessage("there was an error")
            return Promise.reject(res)
        }).then(data => {
            setResponse(data.message)
            setLoading(false)
        }).catch(res => {
            console.log("here")
            console.log(res)
            res.json().then(data => {
                setErrorMessage(<>Code: {res.status}<br/>Message: {data.message}</>)
            })
        })
    }, [])*/
    useEffect(() => {
        fetch("/api/graphql", {
            method: "POST",
            body: JSON.stringify({
                query: "{ tx { hash } }",
            }),
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            setErrorMessage("there was an error")
            return Promise.reject(res)
        }).then(data => {
            console.log(data)
            setResponse(data.message)
            setLoading(false)
        }).catch(res => {
            console.log("here")
            console.log(res)
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
