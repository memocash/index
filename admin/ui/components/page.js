import styles from "../styles/Home.module.css";
import Head from "next/head";
import Link from 'next/link';
import {useEffect, useState} from "react";
import * as config from "./config";

const LocalStorageKey = "server-select"
const SelectValDev = "dev"
const SelectValLive = "live"

export default function Page(props) {
    const [selectValue, setSelectValue] = useState("")

    const setSelect = (val) => {
        setSelectValue(val)
        switch (val) {
            case SelectValDev:
                config.SetHost(config.DevHost)
                break;
            case SelectValLive:
                config.SetHost(config.LiveHost)
                break;
        }
    }

    const selectChange = (e) => {
        setSelect(e.target.value)
        localStorage.setItem(LocalStorageKey, e.target.value)
    }

    useEffect(() => {
        const prevSelect = localStorage.getItem(LocalStorageKey)
        if (prevSelect && prevSelect.length) {
            setSelect(prevSelect)
        }
    })

    return (
        <div className={styles.container}>
            <Head>
                <title>Memo Index Admin</title>
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
                        <li>
                            <Link href="/block/list">
                                <a>Blocks</a>
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
                    <div>
                        <label htmlFor="server-select">Server:</label>
                        &nbsp;
                        <select id={"server-select"} value={selectValue} onChange={selectChange}>
                            <option value={SelectValLive}>Live</option>
                            <option value={SelectValDev}>Dev</option>
                        </select>
                    </div>
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
