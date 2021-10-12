import Page from "../components/page";
import {useEffect, useState} from "react";

function Hello() {
    const [hello, setHello] = useState("")
    const [version, setVersion] = useState("")
    useEffect(() => {
        fetch("/api/hello")
            .then(res => res.json())
            .then(data => {
                setHello(data.Name)
                setVersion(data.Version)
            })
    })
    return (
        <Page>
            <div>
                <h1>
                    Hello Page
                </h1>
                <p>
                    {hello} - {version}
                </p>
            </div>
        </Page>
    )
}

export default Hello
