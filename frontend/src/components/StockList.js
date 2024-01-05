import React, { useState, useEffect } from 'react';
import axios from 'axios';
import './StockList.css';

const StockList = ({ n }) => {
  const [stocks, setStocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [socket, setSocket] = useState(null);

  useEffect(() => {
    const fetchStocks = async () => {
      try {
        const response = await axios.get(`http://localhost:3001/api/stocks?n=${n}`);
        setStocks(response.data);
      } catch (error) {
        console.error('Error fetching stocks:', error.message);
      } finally {
        setLoading(false);
      }
    };

    const initializeWebSocket = () => {
      const ws = new WebSocket('ws://localhost:3001/ws');

      ws.addEventListener('open', () => {
        console.log('WebSocket connected');
      });

      ws.addEventListener('message', (event) => {
        const updatedStocks = JSON.parse(event.data);
        setStocks(updatedStocks);
      });

      ws.addEventListener('close', () => {
        console.log('WebSocket closed. Reconnecting...');
        setTimeout(initializeWebSocket, 1000);
      });

      setSocket(ws);
    };

    initializeWebSocket();
    fetchStocks();

    return () => {
      if (socket) {
        socket.close();
      }
    };
  }, [n, socket]);

  if (loading) {
    return <p>Loading...</p>;
  }

  return (
    <div className="stock-list-container">
      <h2>Stock List</h2>
      <table>
        <thead>
          <tr>
            <th>Symbol</th>
            <th>Open Price</th>
            <th>Current Price</th>
            <th>Refresh Interval</th>
          </tr>
        </thead>
        <tbody>
          {stocks && stocks.length > 0 ? (
            stocks.map((stock) => (
              <tr key={stock.symbol}>
                <td>{stock.symbol}</td>
                <td>{stock.openPrice.toFixed(2)}</td>
                <td>{stock.currentPrice.toFixed(2)}</td>
                <td>{stock.refreshInterval}s</td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan="4">No stocks available</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

export default StockList;
