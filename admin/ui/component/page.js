import styles from "../styles/Home.module.css";
import Head from "next/head";

export default function Page(body) {
    return (
        <div className={styles.container}>
            <Head>
                <title>Memo Server Admin</title>
                <link rel="icon" href="/favicon.ico"/>
            </Head>
            <main className={styles.main}>
                {body}
                <p>
                    <a href="/">Home</a>
                    &middot;
                    <a href="/about">About</a>
                </p>
            </main>
            <footer className={styles.footer}>
            </footer>
        </div>
    )
}
