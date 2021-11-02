import Page from "../../components/page";
import styles from "../../styles/Home.module.css";
import {useEffect, useState} from "react";
import Loading from "../../components/util/loading";
import Link from "next/link";

const query = `
    query {
        double_spends {
            hash
            index
        }
    }
    `

function DoubleSpends() {
    const [doubleSpends, setDoubleSpends] = useState([])
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    useEffect(() => {
        fetch("/api/graphql", {
            method: "POST",
            body: JSON.stringify({
                query: query,
            }),
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            if (data.errors && data.errors.length > 0) {
                setErrorMessage("Double spends graphql errors (see console)")
                setLoading(true)
                console.log(data.errors)
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
                            <div key={doubleSpend}>
                                <Link href={"/tx/" + doubleSpend.hash}>
                                    <a>{doubleSpend.hash}:{doubleSpend.index}</a>
                                </Link>
                            </div>
                        )
                    })}
                </Loading>
            </div>
        </Page>
    )
}

export default DoubleSpends
