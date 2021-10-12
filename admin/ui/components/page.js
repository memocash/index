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
                    <Link href="/peers">
                        <a>Peers</a>
                    </Link>
                    &middot;
                    <Link href="/hello">
                        <a>Hello</a>
                    </Link>
                </p>
            </main>
            <footer className={styles.footer}>
            </footer>
        </div>
    )
}
