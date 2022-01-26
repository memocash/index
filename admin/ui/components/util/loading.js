const Loading = (props) => {
    return (
        <div>
            {props.loading ?
                <div>{!!props.error ?
                    <>Error: <pre>{props.error}</pre></>
                    :
                    <>Loading...</>
                }</div>
                :
                <div>
                    {props.children}
                </div>
            }
        </div>
    )
}

const GetErrorMessage = (errors) => {
    let messages = [];
    for (const id in errors) {
        const error = errors[id]
        let path = ""
        if (error.path && error.path.length) {
            path = "[" + error.path.join(", ") + "]: "
        }
        messages.push(path + error.message)
    }
    return messages.join(", ")
}

module.exports = {
    Loading,
    GetErrorMessage,
}
