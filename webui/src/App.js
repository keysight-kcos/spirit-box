import React, { useState, useEffect } from "react";
import ScriptsDashboard from "./ScriptsDashboard.js";
import UnitDashboard from "./UnitDashboard.js";
import UnitInfo from "./UnitInfo.js";
import TrackerInfo from "./TrackerInfo.js";
import './App.css';

function App() {
	const hostUpEndpoint = `http://${window.location.hostname}:${window.location.port}/host`;
	const quitEndpoint = `http://${window.location.hostname}:${window.location.port}/quit`;
	const [unitInfo, setUnitInfo] = useState({}); // unit who's info is displayed when unitInfo page comes up
	const [unitInfoOpen, setUnitInfoOpen] = useState(false);
	const [hostIsUp, setHostIsUp] = useState(false);

	const [trackerInfo, setTrackerInfo] = useState({}); // info about script runs
	const [trackerInfoOpen, setTrackerInfoOpen] = useState(false);

	useEffect(() => {
		fetch(hostUpEndpoint)
		.then(res => res.text())
		.then(data => {
			if (data !== "not up") {
				setHostIsUp(true);
			} else {
				setInterval(() => {
					if (!hostIsUp) {
						console.log("sending req to hostUp endpoint");
						fetch(hostUpEndpoint)
						.then(res => res.text())
						.then(data => {
							console.log("received:", data);
							if (data !== "not up") {
								setTimeout(() => {
									setHostIsUp(true);
									window.location.reload(true);
								}, 1500);
							}
						})
						.catch(err => console.log(err));
					}
				}, 2500);
			}
		})
		.catch(err => console.log(err));
	}, [hostIsUp, hostUpEndpoint]);

	const handleUnitInfo = (unit) => {
		return (e) => {
			setUnitInfo(unit.Properties);
			setUnitInfoOpen(true);
		};
	};

	const handleTrackerInfo = (tracker) => {
		return (e) => {
			setTrackerInfo(tracker);
			setTrackerInfoOpen(true);
		};
	};

	if (unitInfoOpen) {
		return <UnitInfo unitInfo={unitInfo} close={() => setUnitInfoOpen(false)} />;
	} else if (trackerInfoOpen) {
		return <TrackerInfo tracker={trackerInfo} close={() => setTrackerInfoOpen(false)}/>;
	} else {
		return (
			<div className="bg-blue-300 pl-4 pb-4 h-screen w-full table pr-5">
			<h1 className="text-3xl font-extrabold mb-10 mt-5">
				spirit-box
			</h1>

			<ScriptsDashboard handleTrackerInfo={handleTrackerInfo}/>
			<UnitDashboard handleUnitInfo={handleUnitInfo} />

			<button className="font-bold bg-gray-300 mt-10 p-2 rounded hover:bg-gray-400 shadow-xl" 
			onClick={() => fetch(quitEndpoint)}>

			Shut down spirit-box
			</button>

			</div>
		);
	}
}

export default App;
