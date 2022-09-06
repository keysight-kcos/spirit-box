import React from "react";
import './App.css';

const UnitInfo = ({ unitInfo, close }) => {
	const formatProperty = (property) => {
		if (Array.isArray(property)) {
			return property.join(",\n");
		} else {
			return property;
		}
	};

	return (
		<div className="bg-blue-300 pl-4 pt-5 pb-10">
		<button className="font-bold bg-gray-300 p-2 rounded hover:bg-gray-400 shadow-xl block mb-5" onClick={
			(e) => {
				close();
			}
		}>Back</button>
		<table className="table-auto w-4/5 m-auto shadow-xl ">
			<thead>
			<tr className="tableHeaderRow">
				<th>Property</th>
				<th className="w-1/4-">Value</th>
			</tr>
			</thead>
			<tbody>
				{
					Object.entries(unitInfo).map(p => 
						<tr>
							<td>{p[0]}</td>
							<td>{formatProperty(p[1])}</td>
						</tr>
					)
				}
			</tbody>
		</table>
		</div>
	);
};

export default UnitInfo;
