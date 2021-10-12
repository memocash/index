import Page from "../components/page";
import {useEffect} from "react";

function Peers() {
    useEffect(() => {
        fetch("/api/peers")
            .then(res => res.json())
            .then(data => {
                console.log(data);
            })
    }, [])
    return (
        <Page>
            <div>
                <h1>
                    Peers Page
                </h1>
            </div>
        </Page>
    )
}

export default Peers
