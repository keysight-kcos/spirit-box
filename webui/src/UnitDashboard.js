import React, { useState, useEffect } from "react";
import './App.css';

const UnitDashboard = ({ handleUnitInfo }) => {
	const unitsEndpoint = `http://${window.location.hostname}:${window.location.port}/systemd`;
	const [units, setUnits] = useState([]);
	const [notReady, setNotReady] = useState(0);

	useEffect(() => {
		fetch(unitsEndpoint)
		.then(res => res.json())
		.then(data => {
			console.log("fetched data from systemd endpoint.");
			setUnits(data);
			allReady(data);
		})
		.catch((err) => setUnits([]))
		.then( () => setInterval(
		() => {
			fetch(unitsEndpoint)
			.then(res => res.json())
			.then(data => {
				setUnits(data);
				allReady(data);
			})
			.catch((err) => setUnits([]));
		}, 1000));
	}, [unitsEndpoint]);

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

	const ReadyMessage = () => {
		const isReady = "font-bold rounded-sm bg-emerald-300 p-2 mb-3 w-40 shadow-xl";
		const isNotReady = "font-bold rounded-sm bg-rose-500 p-2 mb-3 w-48 shadow-xl";
		if (notReady === 0) {
			return (
				<div className={isReady}>
					All units are ready.
				</div>
			);
		} else {
			return (
				<div className={isNotReady}>
				{`Waiting for ${notReady} ${unitsPlural()} to be ready...`}
				</div>
			);
		}
	}

	if (units.length === 0) {
		return <div className="notReady">Could not retrieve systemd info from spirit-box.</div>;
	} else {
		return (
			<div className="dashboard w-full mt-10 text-sm">
			<div className="unitContainer">
				<ReadyMessage />
				<table className="table-auto text-left p-10 rounded shadow-xl">
				<thead>
				<tr className="tableHeaderRow">
					<th>Unit</th>
					<th className="pr-5">LoadState</th>
					<th className="pr-5">ActiveState</th>
					<th className="pr-5">SubState</th>
					<th>Observation Time</th>
					<th className="pr-5">Ready Status</th>
				</tr>
				</thead>
				<tbody>
					{units.map(unit => (
						<tr key={unit.Name} className="unitRow" onClick={handleUnitInfo(unit)}>
							<td className="pr-5">{unit.Name}</td>
							<td>{unit.LoadState}</td>
							<td>{unit.ActiveState}</td>
							<td>{unit.SubState}</td>
							<td>{unit.At}</td>
							<ReadyStatus unit={unit}/>
						</tr>
					))}
				</tbody>
				</table>
			</div>
			</div>
		);
	}
};

export default UnitDashboard;
