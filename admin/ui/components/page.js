import styles from "../styles/Home.module.css";
import Head from "next/head";
import Link from 'next/link';
import {useEffect, useState} from "react";
import * as config from "./config";
import {setHost} from "./fetch";

const LocalStorageKey = "server-select"
const SelectValDev = "dev"
const SelectValLive = "live"
const SelectValLiveSv = "live-sv"

export default function Page(props) {
    const [selectValue, setSelectValue] = useState("")
    const [lastHost, setLastHost] = useState("")
    const setSelect = async (val) => {
        setSelectValue(val)
        let host = ""
        switch (val) {
            case SelectValDev:
                host = config.DevHost
                break;
            case SelectValLive:
                host = config.LiveHost
                break;
            case SelectValLiveSv:
                host = config.LiveSvHost
                break;
            default:
                throw "select host value not recognized: " + val
        }
        if (lastHost === "" || host !== lastHost) {
            setHost(host)
        }
        if (lastHost !== "" && host !== lastHost) {
            window.location.reload()
        }
        setLastHost(host)
    }

    const selectChange = async (e) => {
        await setSelect(e.target.value)
        localStorage.setItem(LocalStorageKey, e.target.value)
    }

    useEffect(() => {
        const prevSelect = localStorage.getItem(LocalStorageKey)
        setSelect(prevSelect && prevSelect.length ? prevSelect : SelectValLive)
    }, [])

    return (
        <div className={styles.container}>
            <Head>
                <title>Index</title>
                <link rel="icon" href="/favicon.ico"/>
            </Head>
            <main className={styles.main}>
                <div className={styles.sidebar}>
                    <div className={styles.header}>
                        <h1 className={styles.title}>
                            <Link href="/">
                                Index
                            </Link>
                        </h1>
                    </div>
                    <h3>Explore</h3>
                    <ul>
                        <li>
                            <Link href="/hello">
                                Hello
                            </Link>
                        </li>
                        <li>
                            <Link href="/tx/aa3c2117090349ae08fba883c2f70548b502957ffed8e18c0a5ca8e0b6761cf8">
                                Transaction
                            </Link>
                        </li>
                        <li>
                            <Link href="/address/1Pzdrdoj2NC25GMWknYn18eHYuvLoZ6dpv">
                                Address
                            </Link>
                        </li>
                        <li>
                            <Link href="/tx/double-spends">
                                Double Spends
                            </Link>
                        </li>
                        <li>
                            <Link href="/block/list">
                                Blocks
                            </Link>
                        </li>
                    </ul>
                    <h3>Peer</h3>
                    <ul>
                        <li>
                            <Link href="/tx/broadcast">
                                Broadcast
                            </Link>
                        </li>
                        <li>
                            <Link href="/peer/list">
                                Peer List
                            </Link>
                        </li>
                        <li>
                            <Link href="/peer/report">
                                Peer Report
                            </Link>
                        </li>
                    </ul>
                    <h3>Storage</h3>
                    <ul>
                        <li>
                            <Link href="/topic/list">
                                Topic List
                            </Link>
                        </li>
                    </ul>
                    <div>
                        <label htmlFor="server-select">Server:</label>
                        &nbsp;
                        <select id={"server-select"} value={selectValue} onChange={selectChange}>
                            <option value={SelectValLive}>Live</option>
                            <option value={SelectValLiveSv}>Live SV</option>
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
