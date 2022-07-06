import React, { useState, useEffect } from "react";
import './App.css';

function App() {
	const [messages, setMessages] = useState([]);

	useEffect(() => {
		let ws = new WebSocket(`ws://${window.location.hostname}:8080/socket`);

		ws.onopen = (event) => {
			alert("Connection to websocket established.");
		};

		ws.onmessage = (event) => {
			setMessages(prevMessages => [...prevMessages, event.data]);
		};

		return () => ws.close();
	}, []);

	return (
		<div>
		<div>Messages:</div>
		<ol>
			{messages.map(message => (
				<li key={message}>{message}</li>
			))}
		</ol>
		</div>
	);
}

export default App;
