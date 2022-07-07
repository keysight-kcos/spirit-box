import React, { useState, useEffect } from "react";
import './App.css';

function App() {
	const [units, setUnits] = useState([]);
	const [unitInfo, setUnitInfo] = useState({}); // unit who's info is displayed when unitInfo page comes up

	useEffect(() => {
		let ws = new WebSocket(`ws://${window.location.hostname}:8080/socket`);

		ws.onopen = (event) => {
			alert("Connection to websocket established.");
		};

		ws.onmessage = (event) => {
			setUnits(prevUnits => [...prevUnits, JSON.parse(event.data)]);
		};

		return () => ws.close();
	}, []);

	const makeHandle = (unit) => {
		return (e) => {
			setUnitInfo(unit.Properties);
		};
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
			</tr>
				{units.map(unit => (
					<tr>
						<td>{unit.Name}</td>
						<td>{unit.LoadState}</td>
						<td>{unit.ActiveState}</td>
						<td>{unit.SubState}</td>
						<button onClick={makeHandle(unit)}>
							More Info
						</button>
					</tr>
				))}
			</table>
			</div>
		);
	}
}

export default App;
