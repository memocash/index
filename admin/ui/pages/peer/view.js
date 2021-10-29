import Page from "../../components/page";
import {useEffect, useState} from "react";
import {useRouter} from "next/router";
import styles from '../../styles/list.module.css';
import homeStyles from "../../styles/Home.module.css";

function View() {
    const router = useRouter()
    const [connections, setConnections] = useState([])
    useEffect(() => {
        if (!router || !router.query || !router.query.ip) {
            // Wait until router loaded
            return
        }
        fetch("/api/peer", {
            method: "POST",
            body: JSON.stringify({
                ip: router.query.ip,
                port: router.query.port,
            })
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            console.log(data)
            setConnections(data.Connections)
        }).catch(res => {
            console.log(res)
        })
    }, [router])
    return (
        <Page>
            <div>
                <h2 className={homeStyles.subTitle}>
                    Peer View
                </h2>
                <p>Ip: {router.query.ip}</p>
                <p>Port: {router.query.port}</p>
                <ul className={styles.list}>
                    {connections.map((connection, key) => (
                        <li key={key}>
                            <a>{connection.Ip}:{connection.Port} - {connection.Time} - {connection.Status ? "Success" : "Fail"}</a>
                        </li>
                    ))}
                </ul>
            </div>
        </Page>
    )
}

export default View
