import Page from "../components/page";
import {useEffect, useState} from "react";

function Peers() {
    const [loading, setLoading] = useState(true)
    useEffect(() => {
        fetch("/api/peers")
            .then(res => {
                if (res.ok) {
                    return res.json()
                }
                return Promise.reject(res)
            })
            .then(data => {
                console.log(data);
                setLoading(false)
            })
            .catch(res => {
                setLoading(false)
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Peers Page
                </h1>
                <p>
                    {loading ? "Loading..." : "Loaded"}
                </p>
            </div>
        </Page>
    )
}

export default Peers
