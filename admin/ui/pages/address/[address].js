import styles from '../../styles/Home.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import {Loading, GetErrorMessage} from "../../components/util/loading";
import Link from "next/link";
import {PreInline} from "../../components/util/pre";
import {graphQL} from "../../components/fetch";

export default function LockHash() {
    const router = useRouter()
    const [address, setAddress] = useState({
        address: "",
        balance: 0,
        txs: [],
        profile: {
            name: "",
            profile: "",
            pic: "",
        },
    })
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($address: String!) {
        address(address: $address) {
            address
            txs {
                hash
                inputs {
                    output {
                        lock {
                            address
                        }
                        amount
                    }
                }
                outputs {
                    lock {
                        address
                    }
                    amount
                }
            }
            profile {
                name {
                    name
                    tx_hash
                }
                profile {
                    text
                    tx_hash
                }
                pic {
                    pic
                    tx_hash
                }
            }
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
        graphQL(query, {
            address: address,
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
            for (let i = 0; i < data.data.address.txs.length; i++) {
                const tx = data.data.address.txs[i]
                tx.amount = 0
                for (let j = 0; j < tx.inputs.length; j++) {
                    const input = tx.inputs[j]
                    if (input.output.lock && input.output.lock.address === address) {
                        tx.amount -= input.output.amount
                    }
                }
                for (let j = 0; j < tx.outputs.length; j++) {
                    const output = tx.outputs[j]
                    if (output.lock.address === address) {
                        tx.amount += output.amount
                    }
                }
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
                        <div className={column.width15}>Name</div>
                        <div className={column.width85}>{address.profile.name ?
                            <a href={"/tx/" + address.profile.name.tx_hash}>
                                {address.profile.name.name}
                            </a> : ""}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width50}>
                            <h3>Txs ({address.txs.length})</h3>
                            {address.txs.map((tx, index) => {
                                return (
                                    <div key={index} className={column.container}>
                                        <Link href={"/tx/" + tx.hash}>
                                            <a><PreInline>{tx.hash}</PreInline></a>
                                        </Link>
                                        : {tx.amount}
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
