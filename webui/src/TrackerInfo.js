import React from "react";
import "./App.css";

const TrackerInfo = ({ tracker, close }) => {

	return (
		<div className="bg-blue-300 pl-4 pt-5 pb-10 h-screen">
		<button className="font-bold bg-gray-300 p-2 rounded hover:bg-gray-400 shadow-xl block mb-5" onClick={
			(e) => {
				close();
			}
		}>Back</button>
		<div>Started: <span className="font-bold">{tracker.startTime}</span></div>
		<div>Ended: <span className="font-bold">{tracker.finished ? tracker.endTime : "Not finished"}</span></div>
		<div className="w-4/5 mt-5 m-auto">
		<h1 className="font-bold text-2xl">
			Runs:
		</h1>
		<table className="w-full">
		<thead>
			<tr className="tableHeaderRow">
			<th>
				Info
			</th>
			<th>
				Succeeded
			</th>
			</tr>
		</thead>
		<tbody>
			{tracker.runs.map(run => (
				<tr>
				<td>
				{run.info}
				</td>
				<td>
					{run.success ? "True" : "False"}
				</td>
				</tr>
			))}
		</tbody>
		</table>
		</div>
	</div>
	);
};

export default TrackerInfo;
