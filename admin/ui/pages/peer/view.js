import Page from "../../components/page";
import {useEffect, useState} from "react";
import {useRouter} from "next/router";

function View() {
    const router = useRouter()
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
