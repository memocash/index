import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import Link from "next/link";
import {useRouter} from "next/router";
import {useEffect, useState} from "react";

export default function Topic() {
    const [topic, setTopic] = useState("")
    const router = useRouter()
    let lastTopic = undefined
    useEffect(() => {
        if (!router || !router.query || router.query.topic === lastTopic) {
            return
        }
        const {topic} = router.query
        lastTopic = topic
        setTopic(topic)
    }, [router])

    return (
        <Page>
            <div>
                <h2 className={styles.subTitle}>
                    View Topic
                </h2>
                <p>{topic}</p>
                <p>
                    <Link href={{pathname: "/topic/list"}}>
                        <a>Back to List</a>
                    </Link>
                </p>
            </div>
        </Page>
    )
}
