export default function Loading(props) {
    return (
        <div>
            {props.loading ?
                <p>{!!props.error ?
                    <>Error: {props.error}</>
                    :
                    <>Loading...</>
                }</p>
                :
                <div>
                    {props.children}
                </div>
            }
        </div>
    )
}
