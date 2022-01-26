const pre = require("../../styles/pre.module.css");
const PreInline = (props) => {
    return (
        <pre className={[pre.pre, pre.inline].join(" ")}>
            {props.children}
        </pre>
    )
}

module.exports = {
    PreInline,
}
