import styles from '../../styles/Home.module.css'
import Page from "../../components/page";
import {useRef} from "react";
import {graphQL} from "../../components/fetch";

export default function Broadcast() {
    const rawRef = useRef()
    const onSubmit = (e) => {
        e.preventDefault()
        const query = `mutation ($raw: String!) {
					broadcast(raw: $raw)
				}`
        graphQL(query, {
            raw: rawRef.current.value,
        }).then(res => {
            if (res.ok) {
                return res.json()
            }
            return Promise.reject(res)
        }).then(data => {
            if (data.errors && data.errors.length > 0) {
                console.log(data.errors)
            }
        }).catch(res => {
            console.log(res)
        })
    }
    return (<Page>
        <div>
            <h2 className={styles.subTitle}>
                Broadcast
            </h2>
            <div>
                <form onSubmit={onSubmit}>
                    <label>
                        Raw tx:<br/>
                        <textarea ref={rawRef}/>
                    </label>
                    <br/>
                    {" "}<input type={"submit"} value={"Broadcast"}/>
                </form>
            </div>
        </div>
    </Page>)
}
