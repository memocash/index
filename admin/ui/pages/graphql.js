import styles from '../styles/Home.module.css'
import Page from "../components/page";
import {useState} from "react";
import {graphQL} from "../components/fetch";

const CommonQueries = [
    {
        name: "Address links",
        query: `query ($address: Address!) {
  profiles(addresses: [$address]) {
    address
    links {
      tx_hash
      address
      parent_address
      message
      accepts {
        tx_hash
        request_tx_hash
        message
        revokes {
          tx_hash
          accept_tx_hash
          message
        }
      }
    }
  }
}`,
        variables: `{
  "address": "<address>"
}`,
    },
]

export default function GraphQL() {
    const [query, setQuery] = useState("{\n  \n}")
    const [variables, setVariables] = useState("")
    const [result, setResult] = useState("")
    const [error, setError] = useState("")
    const selectCommon = (e) => {
        const common = CommonQueries[e.target.value]
        if (!common) {
            return
        }
        setQuery(common.query)
        setVariables(common.variables)
    }
    const onSubmit = (e) => {
        e.preventDefault()
        setError("")
        setResult("")
        let parsedVariables = undefined
        const rawVariables = variables.trim()
        if (rawVariables.length) {
            try {
                parsedVariables = JSON.parse(rawVariables)
            } catch (err) {
                setError("Error parsing variables JSON: " + err.message)
                return
            }
        }
        graphQL(query, parsedVariables).then(res => {
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
                    Common queries:{" "}
                    <select value={""} onChange={selectCommon}>
                        <option value={""}>Select a query...</option>
                        {CommonQueries.map((common, i) => (
                            <option key={i} value={i}>{common.name}</option>
                        ))}
                    </select>
                </label>
                <br/>
                <br/>
                <label>
                    Query:<br/>
                    <textarea value={query} onChange={(e) => setQuery(e.target.value)}
                              rows={12} cols={80}/>
                </label>
                <br/>
                <label>
                    Variables (JSON):<br/>
                    <textarea value={variables} onChange={(e) => setVariables(e.target.value)}
                              rows={5} cols={80}/>
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
