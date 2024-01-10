import styles from '../../styles/Home.module.css'
import pre from '../../styles/pre.module.css'
import column from '../../styles/column.module.css'
import Page from "../../components/page";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";
import {GetErrorMessage, Loading} from "../../components/util/loading";
import Link from "next/link";
import {PreInline} from "../../components/util/pre";
import {graphQL} from "../../components/fetch";
import {hexToString} from "../../components/hex";

export default function Hash() {
    const router = useRouter()
    const [tx, setTx] = useState({
        inputs: [],
        outputs: [],
        blocks: [],
    })
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    const query = `
    query ($hash: String!) {
        tx(hash: $hash) {
            hash
            raw
            seen
            version
            locktime
            inputs {
                index
                prev_hash
                prev_index
                script
                output {
                    amount
                    spends {
                        hash
                        index
                    }
                    lock {
                        address
                    }
                    slp {
                        token_hash
                        amount
                        genesis {
                            ticker
                            decimals
                        }
                    }
                    slp_baton {
                        token_hash
                        genesis {
                            ticker
                            decimals
                        }
                    }
                }
            }
            outputs {
                index
                amount
                script
                spends {
                    hash
                    index
                }
                lock {
                    address
                }
                slp {
                    token_hash
                    amount
                    genesis {
                        ticker
                        decimals
                    }
                }
                slp_baton {
                    token_hash
                    genesis {
                        ticker
                        decimals
                    }
                }
            }
            blocks {
                block {
                    hash
                    height
                    timestamp
                }
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
        graphQL(query, {
            hash: hash,
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
            setTx(data.data.tx)
        }).catch(res => {
            setErrorMessage("error loading tx")
            setLoading(true)
            console.log(res)
        })
    }, [router])
    let fee = 0
    for (let i = 0; i < tx.inputs.length; i++) {
        if (tx.inputs[i].output) {
            fee += tx.inputs[i].output.amount
        }
    }
    for (let i = 0; i < tx.outputs.length; i++) {
        fee -= tx.outputs[i].amount
    }
    const size = tx.raw ? tx.raw.length / 2 : 0
    const feeRate = Math.round(fee / size * 1e6) / 1e6
    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    Transaction
                </h2>
                <Loading loading={loading} error={errorMessage}>
                    <div className={column.container}>
                        <div className={column.width15}>Tx hash</div>
                        <div className={column.width85}>{tx.hash}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Tx raw</div>
                        <div className={column.width85}><PreInline>{tx.raw}</PreInline></div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Version</div>
                        <div className={column.width85}>{tx.version}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Locktime</div>
                        <div className={column.width85}>{tx.locktime}</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Size</div>
                        <div className={column.width85}>{size} Bytes</div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>Fee</div>
                        <div className={column.width85}>
                            {hasCoinbase(tx) ? "Coinbase" : (<>{fee} Satoshis ({feeRate} sats/B)</>)}
                        </div>
                    </div>
                    <div className={column.container}>
                        <div className={column.width15}>First Seen</div>
                        <div className={column.width85}>{tx.seen ? tx.seen : "-"}</div>
                    </div>
                    <BlockInfo tx={tx}/>
                    <div className={column.container}>
                        <Inputs tx={tx}/>
                        <Outputs tx={tx}/>
                    </div>
                </Loading>
            </div>
        </Page>
    )
}

function isCoinbase(input) {
    return input.prev_hash === "0000000000000000000000000000000000000000000000000000000000000000"
}

function hasCoinbase(tx) {
    for (let i = 0; i < tx.inputs.length; i++) {
        if (isCoinbase(tx.inputs[i])) {
            return true
        }
    }
    return false
}

function BlockInfo({tx}) {
    return (
        <div className={column.container}>
            <div className={column.width15}>Block</div>
            <div className={column.width85}>
                {tx.blocks ? tx.blocks.map((txBlock) => {
                    return (
                        <div key={txBlock.block.hash}>
                            Hash: <Link href={"/block/" + txBlock.block.hash}>
                            <a>{txBlock.block.hash}</a></Link>
                            <br/>
                            Timestamp: {txBlock.block.timestamp.length ? txBlock.block.timestamp : "Not set"}
                            <br/>
                            Height: {txBlock.block.height}
                        </div>
                    )
                }) : null}
            </div>
        </div>
    )
}

function Inputs({tx}) {
    return (
        <div className={column.width50}>
            <h3>Inputs ({tx.inputs.length})</h3>
            {tx.inputs.map((input) => {
                return (
                    <div key={input.index} className={[column.container, column.marginBottom].join(" ")}>
                        <div className={column.width15}>{input.index}</div>
                        <div className={column.width85}>{input.output ? (<>
                            Address: {input.output.lock ?
                            <Link href={"/address/" + input.output.lock.address}>
                                <a>{input.output.lock.address}</a>
                            </Link>
                            : <>Not found</>}
                            <br/>
                            Amount: {input.output.amount}
                            <br/>
                            UnlockScript: <pre
                            className={[pre.pre, pre.inline].join(" ")}>{input.script}</pre>
                            <br/>
                            {input.output.slp ? <div>
                                SLP: {input.output.slp.amount} {input.output.slp.genesis ?
                                <Link href={"/tx/" + input.output.slp.token_hash}>
                                    <a>
                                        {input.output.slp.genesis.ticker}
                                    </a>
                                </Link> : null}
                            </div> : null}
                            {input.output.slp_baton ? <div>
                                SLP Baton: {input.output.slp_baton.genesis ?
                                <Link href={"/tx/" + input.output.slp_baton.token_hash}>
                                    <a>
                                        {input.output.slp_baton.genesis.ticker}
                                    </a>
                                </Link> : null}
                            </div> : null}
                            {input.output.spends && input.output.spends.length >= 2 ?
                                <div className={[column.red, column.bold].join(" ")}>
                                    DOUBLE SPEND
                                </div>
                                : null
                            }
                            <Link href={"/tx/" + input.prev_hash}>
                                <a><PreInline>{input.prev_hash}:{input.prev_index}</PreInline></a>
                            </Link>
                        </>) : (isCoinbase(input) ? "Coinbase" : (
                            <Link href={"/tx/" + input.prev_hash}>
                                <a><PreInline>{input.prev_hash}:{input.prev_index}</PreInline></a>
                            </Link>
                        ))}</div>
                    </div>
                )
            })}
        </div>
    )
}

const ShortTxHash = (txHash) => {
    if (txHash.length < 16) {
        return txHash
    }
    return txHash.substr(0, 8) + "..." + txHash.substr(txHash.length - 8)
}

const GetOutputScriptInfo = (script) => {
    let info = ""
    if (script.substr(0, 4) === "6a02") {
        switch (script.substr(4, 4)) {
            case "6d01":
                return "Memo name: " + Buffer.from(script.substr(10), "hex")
            case "6d02":
                return "Memo post: " + Buffer.from(script.substr(10), "hex")
            case "6d03":
                const replyTxHash = script.substr(10, 64).match(/.{2}/g).reverse().join("")
                return (<>Memo reply (<Link href={"/tx/" + replyTxHash}><a>{ShortTxHash(replyTxHash)}</a></Link>): {
                    "" + Buffer.from(script.substr(76), "hex")}</>)
            case "6d04":
                if (script.length < 12) {
                    info = "Bad memo like"
                    break
                }
                const likeTxHash = script.substr(10).match(/.{2}/g).reverse().join("")
                return (<>Memo like: <Link href={"/tx/" + likeTxHash}><a>{ShortTxHash(likeTxHash)}</a></Link></>)
            case "6d0a":
                const picUrl = "" + Buffer.from(script.substr(10), "hex")
                return (<>Memo profile pic: <Link href={picUrl}><a>{picUrl}</a></Link></>)
            case "6d0c":
                let size = parseInt(script.substr(8, 2), 16)
                size *= 2
                if (size + 10 > script.length) {
                    info = "Bad topic message"
                    break
                }
                console.log("size:", size)
                return "Memo topic message (" + Buffer.from(script.substr(10, size), "hex") + "): " +
                    Buffer.from(script.substr(10 + size), "hex")
            case "6d0d":
                return "Memo topic follow: " + Buffer.from(script.substr(8), "hex")
            case "6d24":
                return "Memo direct message: " + Buffer.from(script.substr(52), "hex")
            case "6d05":
                return "Memo profile text: " + Buffer.from(script.substr(10), "hex")
        }
    } else if (script.substr(0, 12) === "6a04534c5000") {
        if (hexToString(script.substr(18, 14)) === "GENESIS") {
            return "SLP Genesis"
        } else if (hexToString(script.substr(18, 8)) === "MINT") {
            return "SLP Mint"
        } else if (hexToString(script.substr(18, 8)) === "SEND") {
            return "SLP Send"
        }
    }
    return "Unknown" + (info.length ? ": " + info : "")
}

function Outputs({tx}) {
    return (
        <div className={column.width50}>
            <h3>Outputs ({tx.outputs.length})</h3>
            {tx.outputs.map((output, index) => {
                return (
                    <div key={index} className={[column.container, column.marginBottom].join(" ")}>
                        <div className={column.width15}>
                            {output.index}
                        </div>
                        <div className={column.width85}>
                            {output.lock.address.includes(": ") ? <>{GetOutputScriptInfo(output.script)}</> : <>
                                Address: <Link href={"/address/" + output.lock.address}>
                                <a>{output.lock.address}</a>
                            </Link></>}
                            <br/>
                            Amount: {output.amount}
                            <br/>
                            LockScript: <pre
                            className={[pre.pre, pre.inline].join(" ")}>{output.script}</pre>
                            {output.slp ? <div>
                                SLP: {output.slp.amount} {output.slp.genesis ?
                                <Link href={"/tx/" + output.slp.token_hash}>
                                    <a>
                                        {output.slp.genesis.ticker}
                                    </a>
                                </Link> : null}
                            </div> : null}
                            {output.slp_baton ? <div>
                                SLP Baton: {output.slp_baton.genesis ?
                                <Link href={"/tx/" + output.slp_baton.token_hash}>
                                    <a>
                                        {output.slp_baton.genesis.ticker}
                                    </a>
                                </Link> : null}
                            </div> : null}
                            {output.spends ? <>
                                {output.spends.length >= 2 ?
                                    <div className={[column.red, column.bold].join(" ")}>
                                        DOUBLE SPEND
                                    </div>
                                    : null
                                }
                                <h5 className={column.noMargin}>Spends
                                    ({output.spends.length})</h5>
                                {output.spends.map((spend, index) => {
                                    return (
                                        <div key={index}>
                                            <Link href={"/tx/" + spend.hash}>
                                                <a>
                                                    <PreInline>{spend.hash}:{spend.index}</PreInline>
                                                </a>
                                            </Link>
                                        </div>
                                    )
                                })}
                            </> : null}
                        </div>
                    </div>
                )
            })}
        </div>
    )
}
