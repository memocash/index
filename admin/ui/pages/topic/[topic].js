import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import Link from "next/link";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";

export default function Topic() {
    const [topic, setTopic] = useState("")
    const [topicData, setTopicData] = useState({
        Items: [],
    })
    const router = useRouter()
    let lastTopic = undefined
    useEffect(() => {
        if (!router || !router.query || router.query.topic === lastTopic) {
            return
        }
        lastTopic = router.query.topic
        setTopic(router.query.topic)
        fetch("/api/topic/view", {
            method: "POST",
            body: JSON.stringify({
                topic: router.query.topic,
            }),
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
                <h4>{topic}</h4>
                {topicData.Items && topicData.Items.map((item, key) => {
                    return (
                        <p key={key}>{item.Uid}</p>
                    )
                })}
                <p>
                    <Link href={{pathname: "/topic/list"}}>
                        <a>Back to List</a>
                    </Link>
                </p>
            </div>
        </Page>
    )
}
