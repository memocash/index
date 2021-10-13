import Page from "../../components/page";
import {useEffect, useState} from "react";
import {useRouter} from "next/router";
import styles from '../../styles/list.module.css';
import Link from "next/link";

function View() {
    const router = useRouter()
    const [connections, setConnections] = useState([])
    useEffect(() => {
        fetch("/api/peer", {
            method: "POST",
            body: JSON.stringify({
                Ip: router.query.ip,
                Port: router.query.port,
            })
        })
            .then(res => {
                if (res.ok) {
                    return res.json()
                }
                return Promise.reject(res)
            })
            .then(data => {
                console.log(data)
                setConnections(data.Connections)
            })
            .catch(res => {
                console.log(res)
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Peer Page
                </h1>
                <p>Ip: {router.query.ip}</p>
                <p>Port: {router.query.port}</p>
                <ul className={styles.list}>
                    {connections.map((connection, key) => (
                        <li key={key}>
                            <a>{connection.Ip}:{connection.Port} - {connection.Time} - {connection.Status}</a>
                        </li>
                    ))}
                </ul>
            </div>
        </Page>
    )
}

export default View
