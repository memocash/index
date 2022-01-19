import Page from "../../components/page";
import styles from "../../styles/Home.module.css";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {graphQL} from "../../components/fetch";
import column from "../../styles/column.module.css";
import {useRouter} from "next/router";

const query = `
    query ($start: Date) {
        double_spends(start: $start, newest: True) {
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
                    blocks {
                        hash
                    }
                    seen
                    suspect {
                        hash
                    }
                    lost {
                        hash
                    }
                }
            }
        }
    }
    `

function DoubleSpends() {
    const [doubleSpends, setDoubleSpends] = useState([])
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const [nextStart, setNextStart] = useState("")
    const [lastStart, setLastStart] = useState("")
    const router = useRouter()
    useEffect(() => {
        if (!router || !router.query || (router.query.start === lastStart)) {
            return
        }
        let {start} = router.query
        let variables = {}
        if (start) {
            variables.start = start
        }
        setLastStart(start)
        graphQL(query, variables).then(res => {
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
            if (data.data.double_spends) {
                setNextStart(data.data.double_spends[data.data.double_spends.length - 1].timestamp)
            }
        }).catch((err) => {
            setErrorMessage("Double spends graphql error (see console)")
            console.log(err)
        })
    }, [router])
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
                                                </Link> &middot; ({input.tx.seen})&nbsp;
                                                {input.tx.lost ?
                                                    <span className={[column.red, column.bold].join(" ")}>
                                                        LOST
                                                    </span>
                                                    : (input.tx.suspect ?
                                                        <span className={[column.orange, column.bold].join(" ")}>
                                                            SUSPECT
                                                        </span>
                                                        : "OK!")}
                                                <span> {input.tx.blocks ? "BLOCK" : ""}</span>
                                            </li>
                                        )
                                    })}
                                </ul>
                            </div>
                        )
                    })}
                    <div>
                        <Link href={{
                            pathname: "/tx/double-spends",
                        }}>
                            <a>First</a>
                        </Link>
                        &nbsp;&middot;&nbsp;
                        <Link href={{
                            pathname: "/tx/double-spends",
                            query: {
                                start: nextStart
                            },
                        }}>
                            <a>Next</a>
                        </Link>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}

export default DoubleSpends
