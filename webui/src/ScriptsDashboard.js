import React, { useState, useEffect } from "react";
import Spinner from "./Spinner.js";
import './App.css';

const ScriptsDashboard = ({ handleTrackerInfo }) => {
	const scriptsEndpoint = `http://${window.location.hostname}:8080/scripts`;
	const [priorityGroups, setPriorityGroups] = useState([]);

	useEffect(() => {
		fetch(scriptsEndpoint)
		.then(res => res.json())
		.then(data => {
			console.log("fetched data from scripts endpoint.");
			setPriorityGroups(data);
		})
		.catch((err) => setPriorityGroups([]))
		.then( () => setInterval(
		() => {
			fetch(scriptsEndpoint)
			.then(res => res.json())
			.then(data => {
				setPriorityGroups(data);
			})
			.catch((err) => setPriorityGroups([]));
		}, 1000));
	}, [scriptsEndpoint]);

	const trackerFoundSuccess = (tracker) => tracker.runs[tracker.runs.length-1].success;

	const trackerSuccess = (tracker) => {
		if (trackerFoundSuccess(tracker)) {
			return (
			<td className="ready">
				Succeeded
			</td >);
		} else {
			return (
			<td className="notReady">
				Failed
			</td>
			);
		}
	};

	const ReadyMessage = () => {
		const isReady = "font-bold rounded-sm bg-emerald-300 p-2 mb-3 w-40 shadow-xl";
		const loading = "font-bold rounded-sm bg-amber-500 p-2 mb-3 w-40 shadow-xl";
		const isNotReady = "font-bold rounded-sm bg-rose-500 p-2 mb-3 w-48 shadow-xl";

		let scriptsFinished = 0;
		let scriptsSucceeded = 0;
		let numScripts = 0;

		for (let i = 0; i < priorityGroups.length; i++) {
			let trackers = priorityGroups[i].trackers;
			let specs = priorityGroups[i].specs;
			numScripts += specs.length;
			if (trackers === null) {
				continue;
			}
			for (let j = 0; j < specs.length; j++) {
				if (trackers[j].finished) {
					scriptsFinished++;
					if (trackerFoundSuccess(trackers[j])) {
						scriptsSucceeded++;
					}
				}
			}
		}

		const waiting = `Waiting for ${numScripts - scriptsFinished} scripts to finish...`;
		const successRate = `${scriptsSucceeded}/${numScripts} scripts were successful.`;

		if (scriptsFinished < numScripts) {
			return (
				<div className={loading}>
					{waiting}
				</div>
			);
		} else if (scriptsSucceeded < numScripts) {
			return (
				<div className={isNotReady}>
					{successRate}
				</div>
			);
		} else {
			return (
				<div className={isReady}>
					{successRate}
				</div>
			);
		}
	};

	if (priorityGroups.length === 0) {
		return <div className="notReady mb-5">Could not retrieve script info from spirit-box.</div>;
	} else {
		return (
			<div className="w-full text-sm">
			<ReadyMessage />
			<table>
				<thead>
					<tr className="tableHeaderRow" key="scriptsTableHeader">
						<th>
							Priority Group
						</th>
						<th>
							Command
						</th>
						<th>
							# of Runs
						</th>
						<th>
							Retry Timeout
						</th>
						<th>
							Total Timout	
						</th>
						<th>
							Result
						</th>
					</tr>
				</thead>
				<tbody>
				{priorityGroups.map(pg => (
						<>
						{pg.specs.map((spec, ind) => (
							<tr className="cursor-pointer hover:bg-gray-500"
								key={spec.cmd+spec.args.join()+ind}
								onClick={handleTrackerInfo(pg.trackers === null ? null : pg.trackers[ind])}
							>
								<td>
								{pg.num}
								</td>
								<td>
								{spec.cmd} {spec.args.join(" ")}
								</td>
								<td>
									{pg.trackers === null || pg.trackers[ind] === null ? 0 : pg.trackers[ind].runs.length}
								</td>
								<td>
									{spec.retryTimeout}
								</td>
								<td>
									{spec.totalWaitTime}
								</td>
									{pg.trackers === null
										? <td className="caution">Waiting to execute</td>
										: <Spinner 
											condition={pg.trackers[ind].finished}
											successFunc={() => trackerSuccess(pg.trackers[ind])}
										/>
									}
							</tr>
						))}
						</>
				))}
				</tbody>
			</table>
			</div>
		);
	}
}

export default ScriptsDashboard;
