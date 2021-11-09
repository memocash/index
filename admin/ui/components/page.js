import styles from "../styles/Home.module.css";
import Head from "next/head";
import Link from 'next/link';

export default function Page(props) {
    return (
        <div className={styles.container}>
            <Head>
                <title>Memo Server Admin</title>
                <link rel="icon" href="/favicon.ico"/>
            </Head>
            <main className={styles.main}>
                <div className={styles.sidebar}>
                    <div className={styles.header}>
                        <h1 className={styles.title}>
                            <Link href="/">
                                <a>
                                    Memo Admin
                                </a>
                            </Link>
                        </h1>
                    </div>
                    <ul>
                        <li>
                            <Link href="/hello">
                                <a>Hello</a>
                            </Link>
                        </li>
                        <li>
                            <Link href="/tx/aa3c2117090349ae08fba883c2f70548b502957ffed8e18c0a5ca8e0b6761cf8">
                                <a>Transaction</a>
                            </Link>
                        </li>
                        <li>
                            <Link href="/address/1Pzdrdoj2NC25GMWknYn18eHYuvLoZ6dpv">
                                <a>Address</a>
                            </Link>
                        </li>
                        <li>
                            <Link href="/tx/double-spends">
                                <a>Double Spends</a>
                            </Link>
                        </li>
                    </ul>
                    <h3>Peer</h3>
                    <ul>
                        <li>
                            <Link href="/peer/list">
                                <a>Peer List</a>
                            </Link>
                        </li>
                        <li>
                            <Link href="/peer/report">
                                <a>Peer Report</a>
                            </Link>
                        </li>
                    </ul>
                </div>
                <div className={styles.content}>
                    {props.children}
                </div>
            </main>
            <footer className={styles.footer}>
            </footer>
        </div>
    )
}
