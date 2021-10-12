import Page from "../../components/page";
import {useEffect, useState} from "react";
import {useRouter} from "next/router";

function View() {
    const router = useRouter()
    useEffect(() => {
        fetch("/api/peer", {
            method: "POST",
            body: JSON.stringify({
                Ip: router.query.ip,
                Port: router.query.port,
            })
        })
            .then(res => {
                if (res.ok) {
                    return res.json()
                }
                return Promise.reject(res)
            })
            .then(data => {
                console.log(data)
            })
            .catch(res => {
                console.log(res)
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Peer Page
                </h1>
                <p>Get params</p>
                <p>Ip: {router.query.ip}</p>
                <p>Port: {router.query.port}</p>
            </div>
        </Page>
    )
}

export default View
