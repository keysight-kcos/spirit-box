import React, { useState, useEffect, useCallback } from "react";
import './App.css';

function App() {
	const [units, setUnits] = useState([]);
	const [unitInfo, setUnitInfo] = useState({}); // unit who's info is displayed when unitInfo page comes up
	const [connected, setConnected] = useState(false);
	const [socket, setSocket] = useState({});

	const connect = useCallback(() => {
		let ws = new WebSocket(`ws://${window.location.hostname}:8080/socket`);

		ws.onopen = (event) => {
			setConnected(true);
		};

		ws.onclose = (event) => {
			setConnected(false);
			setTimeout(() => {
				connect();
			}, 1000);
		};

		ws.onmessage = (event) => {
			setUnits(prevUnits => [...prevUnits, JSON.parse(event.data)]);
		}

		return ws;

	}, []);

	useEffect(() => {
		let ws = connect();
		setSocket(ws);
		return () => ws.close();
	}, [connect]);

	const makeHandle = (unit) => {
		return (e) => {
			setUnitInfo(unit.Properties);
		};
	};

	const ConnStatus =  () => connected ? <div>Connected</div> : <div>Not connected</div>;

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
			<ConnStatus/>
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
			<button onClick={() => socket.send("stop")}>
				Stop spirit-box
			</button>
			</div>
		);
	}
}

export default App;
