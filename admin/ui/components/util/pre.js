const pre = require("../../styles/pre.module.css");
exports.PreInline = function (props) {
    return (
        <pre className={[pre.pre, pre.inline].join(" ")}>
            {props.children}
        </pre>
    )
}
