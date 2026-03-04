import styles from '../styles/Home.module.css'
import column from '../styles/column.module.css'
import Link from "next/link";
import {Loading} from "./util/loading";

const PAGE_SIZE = 25

export default function BlockDetail({block, lastOffset, offset, hashLink, txPaginationPath, loading, errorMessage}) {
    const prevOffset = Math.max(0, lastOffset - PAGE_SIZE)
    return (
        <div>
            <h2 className={styles.subTitle}>
                Block
            </h2>
            <Loading loading={loading} error={errorMessage}>
                <div className={column.container}>
                    <div className={column.width15}>Hash</div>
                    <div className={column.width85}>
                        {hashLink ?
                            <Link href={"/block/" + block.hash}>
                                {block.hash}
                            </Link>
                            : block.hash
                        }
                    </div>
                </div>
                <div className={column.container}>
                    <div className={column.width15}>Timestamp</div>
                    <div className={column.width85}>{block.timestamp}</div>
                </div>
                <div className={column.container}>
                    <div className={column.width15}>Height</div>
                    <div className={column.width85}>{block.height.toLocaleString()}</div>
                </div>
                <div className={column.container}>
                    <div className={column.width15}>Raw</div>
                    <div className={column.width85}>
                        <pre className={column.pre}>{block.raw}</pre>
                    </div>
                </div>
                <div className={column.container}>
                    <div className={column.width15}>Size</div>
                    <div className={column.width85}>{block.size ? block.size.toLocaleString() : 0} bytes</div>
                </div>
                <div className={column.navButtons}>
                    {block.height > 0 &&
                        <Link href={"/block/height/" + (block.height - 1)}>
                            <span className={column.navButton}>&laquo; {(block.height - 1).toLocaleString()}</span>
                        </Link>
                    }
                    {" "}
                    <Link href={"/block/height/" + (block.height + 1)}>
                        <span className={column.navButton}>{(block.height + 1).toLocaleString()} &raquo;</span>
                    </Link>
                </div>
                <div className={column.container}>
                    <div>{block.txs && block.txs.length > 0 ? <>
                        <h3>Txs ({(lastOffset + 1).toLocaleString()}&ndash;{(lastOffset + block.txs.length).toLocaleString()} of
                            {" " + (block.tx_count ? block.tx_count.toLocaleString() : block.txs.length)})</h3>
                        <table className={column.container}>
                            <tbody>
                            {block.txs.map((txBlock) => {
                                return (
                                    <tr key={txBlock.index}>
                                        <td>{txBlock.index}.</td>
                                        <td>
                                            <Link href={"/tx/" + txBlock.tx.hash}>
                                                {txBlock.tx.hash}
                                            </Link>
                                        </td>
                                    </tr>
                                )
                            })}
                            </tbody>
                        </table>
                    </> : <>No transactions</>}
                    </div>
                </div>
                {block.txs && block.txs.length > 0 && <div className={column.navButtons}>
                    {lastOffset > 0 ?
                        <Link href={{
                            pathname: txPaginationPath,
                            query: prevOffset > 0 ? {start: prevOffset} : undefined,
                        }}>
                            <span className={column.navButton}>&laquo; Back</span>
                        </Link>
                        : <span className={column.navButtonDisabled}>&laquo; Back</span>
                    }
                    {" "}
                    {offset < block.tx_count ?
                        <Link href={{
                            pathname: txPaginationPath,
                            query: {start: offset},
                        }}>
                            <span className={column.navButton}>Next &raquo;</span>
                        </Link>
                        : <span className={column.navButtonDisabled}>Next &raquo;</span>
                    }
                </div>}
            </Loading>
        </div>
    )
}
