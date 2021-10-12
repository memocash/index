import Page from "../components/page";
import {useEffect, useState} from "react";

function Hello() {
    const [hello, setHello] = useState("")
    const [version, setVersion] = useState("")
    const [loading, setLoading] = useState(true)
    const [errorMessage, setErrorMessage] = useState("")
    useEffect(() => {
        fetch("/api/hello")
            .then(res => {
                if (res.ok) {
                    return res.json()
                }
                setErrorMessage("there was an error")
                return Promise.reject(res)
            })
            .then(data => {
                setHello(data.Name)
                setVersion(data.Version)
                setLoading(false)
            })
            .catch(res => {
                res.json()
                    .then(data => {
                        setErrorMessage(<>Code: {res.status}<br/>Message: {data.message}</>)
                    })
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Hello Page
                </h1>
                <p>{loading ?
                    <>{!!errorMessage ?
                        <>Error: {errorMessage}</>
                        :
                        <>Loading...</>
                    }</>
                    :
                    <>{hello} - {version}</>
                }</p>
            </div>
        </Page>
    )
}

export default Hello
