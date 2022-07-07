import React, { useState, useEffect } from "react";
import './App.css';

function App() {
	const unitsEndpoint = `http://${window.location.hostname}:8080/systemd`;
	const quitEndpoint = `http://${window.location.hostname}:8080/quit`;
	const [units, setUnits] = useState([]);
	const [unitInfo, setUnitInfo] = useState({}); // unit who's info is displayed when unitInfo page comes up
	const [notReady, setNotReady] = useState(0);

	useEffect(() => {
		setInterval(
		() => {
			fetch(unitsEndpoint)
			.then(res => res.json())
			.then(data => {
				setUnits(data);
				allReady(data);
			})
			.catch((err) => setUnits([]));
		}, 1000);
	}, [unitsEndpoint]);

	const makeHandle = (unit) => {
		return (e) => {
			setUnitInfo(unit.Properties);
		};
	};

	const ReadyStatus = ({ unit }) => {
		//console.log("rendering", unit.Name);
		let s = "NOT READY";
		let style = "notReady";
		if (unit.Ready) {
			s = "READY";
			style = "ready";
		} 
		if (unit.SubStateDesired === "watch") {
			s = "WATCHING";
		}

		return (
			<td className={style}>
				{s}
			</td>
		);
	};

	const allReady = (units) => {
		let notReady = 0;
		for (let i = 0; i < units.length; i++)  {
			const unit = units[i];
			if (!unit.Ready) {
				notReady++;
			}
		}
		setNotReady(notReady);
	};

	const unitsPlural = () => {
		return notReady > 1 
			? "units"
			: "unit";
	};

	const UnitDashboard = () => {
		if (units.length === 0) {
			return <div className="noConnection">Could not connect to spirit-box.</div>;
		} else {
			return (
				<div className="dashboard">
				<div className="unitContainer">
					<div>
					{notReady === 0 
						? "All units are ready." 
						: `Waiting for ${notReady} ${unitsPlural()} to be ready...`}
					</div>
					<table>
					<tr className="tableHeaderRow">
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
				</div>
			);
		}
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
			<>
			<h1 className="title">
				spirit-box
			</h1>
			<UnitDashboard />
			<button className="quitButton" onClick={() => fetch(quitEndpoint)}>
			Shut down spirit-box
			</button>
			</>
		);
	}
}

export default App;
