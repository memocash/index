import Page from "../../components/page";
import homeStyles from "../../styles/Home.module.css";
import {useEffect, useState} from "react";

function List() {
    const [allTopics, setAllTopics] = useState([])

    useEffect(() => {
        fetch("/api/topic/list").then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            console.log(data)
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
                <div>
                    {allTopics.map((topic) => {
                        return (
                            <p>{topic.Name}</p>
                        )
                    })}
                </div>
            </div>
        </Page>
    )
}

export default List
