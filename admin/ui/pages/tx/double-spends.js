import Page from "../../components/page";
import styles from "../../styles/Home.module.css";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {graphQL} from "../../components/fetch";

const query = `
    query {
        double_spends {
            hash
            index
            output {
                tx {
                    seen
                }
            }
            inputs {
                hash
                index
                tx {
                    seen
                }
            }
        }
    }
    `

function DoubleSpends() {
    const [doubleSpends, setDoubleSpends] = useState([])
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    useEffect(() => {
        graphQL(query).then(res => {
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
            setDoubleSpends(data.data.double_spends)
        }).catch((err) => {
            setErrorMessage("Double spends graphql error (see console)")
            console.log(err)
        })
    }, [])
    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    Double Spends
                </h2>
                <h3>
                    Outputs ({doubleSpends.length})
                </h3>
                <Loading loading={loading} error={errorMessage}>
                    {doubleSpends.map((doubleSpend) => {
                        return (
                            <div key={doubleSpend.hash + doubleSpend.index}>
                                <Link href={"/tx/" + doubleSpend.hash}>
                                    <a>{doubleSpend.hash}:{doubleSpend.index}</a>
                                </Link> &middot; ({doubleSpend.output.tx.seen})
                                <ul>
                                    {doubleSpend.inputs.map((input) => {
                                        return (
                                            <li key={input.hash + input.index}>
                                                <Link href={"/tx/" + input.hash}>
                                                    <a>{input.hash}:{input.index}</a>
                                                </Link> &middot; ({input.tx.seen})
                                            </li>
                                        )
                                    })}
                                </ul>
                            </div>
                        )
                    })}
                </Loading>
            </div>
        </Page>
    )
}

export default DoubleSpends
