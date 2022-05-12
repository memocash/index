import styles from '../../../styles/Home.module.css'
import Page from "../../../components/page";
import Link from "next/link";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";

export default function Topic() {
    const [topic, setTopic] = useState("")
    const [shard, setShard] = useState(undefined)
    const [uid, setUid] = useState(undefined)
    const [item, setItem] = useState({Message: ""})
    const router = useRouter()
    let lastTopic = undefined
    let lastUid = undefined
    let lastShard = undefined
    useEffect(() => {
        if (!router || !router.query || (router.query.topic === lastTopic && router.query.shard === lastShard &&
            router.query.uid === lastUid)) {
            return
        }
        lastTopic = router.query.topic
        lastShard = router.query.shard
        lastUid = router.query.uid
        setTopic(router.query.topic)
        setShard(router.query.shard)
        setUid(router.query.uid)
        fetch("/api/topic/item", {
            method: "POST",
            body: JSON.stringify({
                topic: router.query.topic,
                shard: router.query.shard,
                uid: router.query.uid,
            }),
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            setItem(data.Item)
        }).catch(err => {
            console.log(err)
        })
    }, [router])
    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    View Item
                </h2>
                <h4>
                    <Link href={{pathname: "/topic/" + topic}}>
                        <a>{topic}</a>
                    </Link>
                </h4>
                <p>{uid}</p>
                <div>
                    {item.Message}
                </div>
                <p>
                    <Link href={{pathname: "/topic/list"}}>
                        <a>Back to List</a>
                    </Link>
                </p>
            </div>
        </Page>
    )
}
