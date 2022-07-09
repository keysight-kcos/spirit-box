import React, { useState, useEffect } from "react";
import './App.css';

const ScriptsDashboard = ({ priorityGroups }) => {

	const TrackerInfo = ({ tracker }) => {
		return (
		<div className="runsContainer">
			Runs:
			<div>Started: {tracker.startTime}</div>
			<div>Ended: {tracker.finished ? tracker.endTime : "Not finished"}</div>
				{tracker.runs.map(run => (
					<div>
						success: {run.success ? "True" : "False"}, info: {run.info}
					</div>
				))}
		</div>
		);
	};

	const ScriptInfo = ({ specs, trackers })  => {
		return (
			<div className="specInfo">
			{specs.map((spec, ind) => (
				<div className="commandContainer">
				<div>Command: {spec.cmd} {spec.args.map(arg => `${arg} `)}</div>
				<div>Timeout between retries: {spec.retryTimeout}ms</div>
				<div>Total time allowed on retries: {spec.totalWaitTime}ms</div>
				{(trackers === null || trackers[ind] === null)
				?  <div>Waiting for script to be scheduled.</div>
				: <TrackerInfo tracker={trackers[ind]} />
				}
				</div>
			))}
			</div>
		);
	};

	if (priorityGroups.length === 0) {
		return <div className="noConnection">Could not retrieve script info from spirit-box.</div>;
	} else {
		return (
			<div>
			{priorityGroups.map(pg => (
				<>
				<div className="pgContainer">
					Priority Group {pg.num}:
				<ScriptInfo specs={pg.specs} trackers={pg.trackers} />
				</div>
				</>
			))}
			</div>
		);
	}
}

export default ScriptsDashboard;
