import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import Link from "next/link";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";

export default function Topic() {
    const [topic, setTopic] = useState("")
    const [shard, setShard] = useState(undefined)
    const [start, setStart] = useState(undefined)
    const [topicData, setTopicData] = useState({
        Items: [],
    })
    const router = useRouter()
    let lastTopic = undefined
    let lastShard = undefined
    let lastStart = undefined
    useEffect(() => {
        if (!router || !router.query || (router.query.topic === lastTopic && router.query.shard === lastShard &&
            router.query.start === lastStart)) {
            return
        }
        lastTopic = router.query.topic
        lastShard = router.query.shard
        lastStart = router.query.start
        setTopic(router.query.topic)
        setShard(router.query.shard)
        setStart(router.query.start)
        let data = {
            topic: router.query.topic,
            shard: router.query.shard,
        }
        if (data.shard === undefined) {
            data.shard = -1
        }
        if (router.query.start !== undefined) {
            data.start = router.query.start
        }
        fetch("/api/topic/view", {
            method: "POST",
            body: JSON.stringify(data),
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            setTopicData(data)
        }).catch(err => {
            console.log(err)
        })
    }, [router])
    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    View Topic
                </h2>
                <h4>
                    <Link href={{pathname: "/topic/" + topic}}>
                        <a>{topic}</a>
                    </Link>
                </h4>
                <div>
                    <label>Start UID: <input type={"text"}/></label>
                    &nbsp;&nbsp;
                    <label>Shard (empty=all): <input type={"number"}/></label>
                </div>
                {topicData.Items && topicData.Items.map((item, key) => {
                    return (
                        <p key={key}>{item.Shard}: {item.Uid}</p>
                    )
                })}
                <p>
                    <Link href={{pathname: "/topic/list"}}>
                        <a>Back to List</a>
                    </Link>
                    {topicData.Items && topicData.Items.length &&
                    <Link href={{
                        pathname: "/topic/" + topic, query: {
                            shard: topicData.Items[topicData.Items.length - 1].Shard,
                            start: topicData.Items[topicData.Items.length - 1].Uid,
                        }
                    }}>
                        <a>Next Page</a>
                    </Link>
                    }
                </p>
            </div>
        </Page>
    )
}
