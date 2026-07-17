import styles from '../styles/Home.module.css'
import Page from "../components/page";
import {useRef, useState} from "react";
import {graphQL} from "../components/fetch";

export default function GraphQL() {
    const queryRef = useRef()
    const variablesRef = useRef()
    const [result, setResult] = useState("")
    const [error, setError] = useState("")
    const onSubmit = (e) => {
        e.preventDefault()
        setError("")
        setResult("")
        let variables = undefined
        const rawVariables = variablesRef.current.value.trim()
        if (rawVariables.length) {
            try {
                variables = JSON.parse(rawVariables)
            } catch (err) {
                setError("Error parsing variables JSON: " + err.message)
                return
            }
        }
        graphQL(queryRef.current.value, variables).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            setResult(JSON.stringify(data, null, 2))
        }).catch(res => {
            setError("Error running query")
            console.log(res)
        })
    }
    return (<Page>
        <div>
            <h2 className={styles.subTitle}>
                GraphQL
            </h2>
            <form onSubmit={onSubmit}>
                <label>
                    Query:<br/>
                    <textarea ref={queryRef} rows={12} cols={80}
                              defaultValue={"{\n  \n}"}/>
                </label>
                <br/>
                <label>
                    Variables (JSON):<br/>
                    <textarea ref={variablesRef} rows={5} cols={80}/>
                </label>
                <br/>
                {" "}<input type={"submit"} value={"Run"}/>
            </form>
            {error.length > 0 ?
                <p>{error}</p> : null}
            {result.length > 0 ?
                <pre style={{whiteSpace: "pre-wrap", overflowWrap: "anywhere"}}>{result}</pre> : null}
        </div>
    </Page>)
}
