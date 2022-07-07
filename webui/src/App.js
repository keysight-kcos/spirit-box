import React, { useState, useEffect, useCallback } from "react";
import './App.css';

function App() {
	const unitsEndpoint = `http://${window.location.hostname}:8080/systemd`;
	const [units, setUnits] = useState([]);
	const [unitInfo, setUnitInfo] = useState({}); // unit who's info is displayed when unitInfo page comes up

	useEffect(() => {
		setInterval(
		() => {
			fetch(unitsEndpoint)
			.then(res => res.json())
			.then(data => setUnits(data))
		}, 1000);
	}, []);

	const makeHandle = (unit) => {
		return (e) => {
			setUnitInfo(unit.Properties);
		};
	};

	const ReadyStatus = (unit) => {
		let s = "NOT READY";
		let style = "notReady";
		if (unit.SubStateDesired === "any" || unit.SubStateDesired === unit.SubState) {
			s = "READY";
			style = "ready";
		} else if (unit.SubStateDesired === "watch") {
			s = "WATCHING";
			style = "ready";
		}

		return (
			<td className={style}>
				{s}
			</td>
		);
	};

	if (Object.keys(unitInfo).length !== 0) {
		console.log(unitInfo);
		return (
			<div>
			<button onClick={
				(e) => {
					setUnitInfo({});
				}
			}>Back</button>
			<ul>
				{
					Object.entries(unitInfo).map(p => {
						return <li>{`${p[0]}: ${p[1]}`}</li>;
					})
				}
			</ul>
			</div>
		);
	} else {
		return (
			<div>
			<table>
			<tr>
				<th>Unit</th>
				<th>LoadState</th>
				<th>ActiveState</th>
				<th>SubState</th>
				<th>Ready Status</th>
				<th>Observation Time</th>
			</tr>
				{units.map(unit => (
					<tr className="unitRow" onClick={makeHandle(unit)}>
						<td>{unit.Name}</td>
						<td>{unit.LoadState}</td>
						<td>{unit.ActiveState}</td>
						<td>{unit.SubState}</td>
						<ReadyStatus unit={unit}/>
						<td>{unit.At}</td>
					</tr>
				))}
			</table>
			</div>
		);
	}
}

export default App;
