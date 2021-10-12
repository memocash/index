import Page from "../components/page";
import {useEffect, useState} from "react";

function Hello() {
    const [hello, setHello] = useState("")
    const [version, setVersion] = useState("")
    const [loading, setLoading] = useState(true)
    useEffect(() => {
        fetch("/api/hello")
            .then(res => res.json())
            .then(data => {
                setHello(data.Name)
                setVersion(data.Version)
                setLoading(false)
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Hello Page
                </h1>
                {loading ?
                    <p>Loading...</p>
                    :
                    <p>{hello} - {version}</p>
                }
            </div>
        </Page>
    )
}

export default Hello
