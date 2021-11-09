import styles from '../../styles/Home.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import {Loading, GetErrorMessage} from "../../components/util/loading";

export default function LockHash() {
    const router = useRouter()
    const [address, setAddress] = useState({
        address: "",
        balance: 0,
    })
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($address: String!) {
        address(address: $address) {
            hash
            address
            balance
        }
    }
    `
    let lastAddress = undefined
    useEffect(() => {
        if (!router || !router.query || router.query.address === lastAddress) {
            return
        }
        const {address} = router.query
        lastAddress = address
        fetch("/api/graphql", {
            method: "POST",
            body: JSON.stringify({
                query: query,
                variables: {
                    address: address,
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
            setAddress(data.data.address)
        }).catch(res => {
            setErrorMessage("error loading address")
            setLoading(true)
            console.log(res)
        })
    }, [router])

    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    Address
                </h2>
                <Loading loading={loading} error={errorMessage}>
                    <div className={column.container}>
                        <div className={column.width15}>Address</div>
                        <div className={column.width85}>{address.address}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Balance</div>
                        <div className={column.width85}>{address.balance.toLocaleString()}</div>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}
