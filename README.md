# Stock Monitoring App

This is a simple stock monitoring application that fetches stock data from the Polygon.io API, updates the data at regular intervals, and provides real-time updates to connected clients via WebSocket.

## Features

- Fetches stock data from Polygon.io API.
- Uses WebSocket for real-time updates to connected clients.
- Provides a RESTful API for accessing stock data.

## Prerequisites

- Go (v1.16 or higher)
- Node.js and npm (for the frontend application)
- Polygon.io API key (Get it [here](https://polygon.io/))

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/stock-monitoring-app.git
   cd stock-monitoring-app


1. Set up environment variables:

Create a .env file in the project root and add the following:

POLYGON_API_KEY=your_polygon_api_key_here

2. Install dependencies:

go mod tidy
npm install --prefix frontend

3. Build the frontend:

npm run build --prefix frontend

4. Run the application:

go run main.go

The application will be accessible at http://localhost:3001.
