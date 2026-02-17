import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import ErrorBoundary from "./components/common/ErrorBoundary";
import { loadConfig } from "./env";
import "@xterm/xterm/css/xterm.css";
import "font-logos/assets/font-logos.css";
import "./index.css";

loadConfig().then(() => {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <ErrorBoundary>
        <BrowserRouter basename="/v2/ui">
          <App />
        </BrowserRouter>
      </ErrorBoundary>
    </React.StrictMode>,
  );
});
