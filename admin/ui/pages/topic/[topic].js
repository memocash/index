import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import Link from "next/link";
import {useRouter} from "next/router";
import {useEffect, useRef, useState} from "react";
import {getUrl} from "../../components/fetch";

export default function Topic() {
    const [topic, setTopic] = useState("")
    const startRef = useRef()
    const shardRef = useRef()
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
        if (router.query.start && router.query.start.length) {
            startRef.current.value = router.query.start
        }
        if (router.query.shard && router.query.shard.length) {
            shardRef.current.value = router.query.shard
        }
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
        getUrl("/api/topic/view", {
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
    const formSubmit = (e) => {
        e.preventDefault()
        let query = {}
        if (startRef.current.value.length) {
            query.start = startRef.current.value
        }
        if (shardRef.current.value.length) {
            query.shard = shardRef.current.value
        }
        router.push({
            pathname: "/topic/" + topic,
            query: query,
        })
    }
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
                <form onSubmit={formSubmit}>
                    <label>Start UID: <input type={"text"} ref={startRef}/></label>
                    &nbsp;&nbsp;
                    <label>Shard (empty=all): <input type={"number"} ref={shardRef}/></label>
                    &nbsp;
                    <input type={"submit"} value={"Update"}/>
                </form>
                {topicData.Items && topicData.Items.map((item, key) => (
                    <p key={key}>{item.Shard}: <Link href={{
                        pathname: "/topic/" + topic + "/" + item.Uid,
                        query: {shard: item.Shard},
                    }}><a>{item.Uid}</a></Link></p>
                ))}
                <p>
                    <Link href={{pathname: "/topic/list"}}>
                        <a>Back to List</a>
                    </Link> &middot; {topicData.Items && topicData.Items.length &&
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
