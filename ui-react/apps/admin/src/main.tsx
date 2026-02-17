import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import "@xterm/xterm/css/xterm.css";
import "font-logos/assets/font-logos.css";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter basename="/v2/ui">
      <App />
    </BrowserRouter>
  </React.StrictMode>,
);
