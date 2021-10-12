export default function BoilingVerdict(props) {
    if (props.celsius >= 100) {
        return <p>The water WOULD boil.</p>
    } else {
        return <p>The water WOULD NOT boil.</p>
    }
}
