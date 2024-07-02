import Page from "../../components/page";
import homeStyles from "../../styles/Home.module.css";
import {useEffect, useState} from "react";
import Link from "next/link";
import {getUrl} from "../../components/fetch";

function List() {
    const [allTopics, setAllTopics] = useState([])

    useEffect(() => {
        getUrl("/api/topic/list").then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            setAllTopics(data.Topics)
        }).catch(err => {
            console.log(err)
        })
    }, [])
    return (
        <Page>
            <div>
                <h2 className={homeStyles.subTitle}>
                    Topic List
                </h2>
                <ul>
                    {allTopics.map((topic, key) => {
                        return (
                            <li key={key}>
                                <Link href={{pathname: "/topic/" + topic.Name}}>
                                    {topic.Name}
                                </Link>
                            </li>
                        )
                    })}
                </ul>
            </div>
        </Page>
    )
}

export default List
