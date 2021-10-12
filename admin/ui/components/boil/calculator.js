import BoilingVerdict from "./boiling";
import {useState} from "react";
import TemperatureInput from "./temperature";

function toCelsius(fahrenheit) {
    return (fahrenheit - 32) * 5 / 9;
}

function toFahrenheit(celsius) {
    return (celsius * 9 / 5) + 32;
}

function tryConvert(temperature, convert) {
    const input = parseFloat(temperature);
    if (Number.isNaN(input)) {
        return '';
    }
    const output = convert(input);
    const rounded = Math.round(output * 1000) / 1000;
    return rounded.toString();
}

export default function Calculator() {
    const [celsius, setCelsius] = useState("");
    const [fahrenheit, setFahrenheit] = useState("");
    const updateCelsius = e => {
        setCelsius(e.target.value)
        setFahrenheit(tryConvert(e.target.value, toFahrenheit))
    }
    const updateFahrenheit = e => {
        setFahrenheit(e.target.value)
        setCelsius(tryConvert(e.target.value, toCelsius))
    }
    return (
        <div>
            <TemperatureInput
                scale="c"
                temperature={celsius}
                updateTemperature={updateCelsius}
            />
            <TemperatureInput
                scale="f"
                temperature={fahrenheit}
                updateTemperature={updateFahrenheit}
            />
            <BoilingVerdict
                celsius={parseFloat(celsius)}
            />
        </div>
    )
}
