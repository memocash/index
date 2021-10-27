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
                {props.children}
                <p>
                    <Link href="/">
                        <a>Home</a>
                    </Link>
                    &middot;
                    <Link href="/peer/list">
                        <a>Peer List</a>
                    </Link>
                    &middot;
                    <Link href="/peer/report">
                        <a>Peer Report</a>
                    </Link>
                    &middot;
                    <Link href="/hello">
                        <a>Hello</a>
                    </Link>
                    &middot;
                    <Link href="/tx/aa3c2117090349ae08fba883c2f70548b502957ffed8e18c0a5ca8e0b6761cf8">
                        <a>Transaction</a>
                    </Link>
                </p>
            </main>
            <footer className={styles.footer}>
            </footer>
        </div>
    )
}
