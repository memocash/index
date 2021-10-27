export default function Loading(props) {
    return (
        <div>
            {props.loading ?
                <div>{!!props.error ?
                    <>Error: {props.error}</>
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
