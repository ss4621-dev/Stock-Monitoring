// src/App.js
import React, { useState } from 'react';
import StockList from './components/StockList';

const App = () => {
  const [n, setN] = useState(10);

  const handleInputChange = (event) => {
    const value = parseInt(event.target.value, 10);
    setN(value > 20 ? 20 : value);
  };

  return (
    <div>
      <h1>Stock Market App</h1>
      <label htmlFor="stockCount">Number of Stocks (max 20): </label>
      <input
        type="number"
        id="stockCount"
        name="stockCount"
        min="1"
        max="20"
        value={n}
        onChange={handleInputChange}
      />
      <StockList n={n} />
    </div>
  );
};

export default App;
