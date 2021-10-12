import {useState} from "react";

const ScaleNames = {
    c: "Celsius",
    f: "Fahrenheit"
}

export default function TemperatureInput(props) {
    return (
        <fieldset>
            <legend>Enter temperature in {ScaleNames[props.scale]}:</legend>
            <input
                value={props.temperature}
                onChange={props.updateTemperature}/>
        </fieldset>
    )
}
