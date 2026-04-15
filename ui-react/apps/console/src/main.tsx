import React from "react";
import ReactDOM from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import ErrorBoundary from "./components/common/ErrorBoundary";
import { ClipboardProvider } from "./components/common/ClipboardProvider";
import { loadConfig } from "./env";
import { queryClient } from "./api/queryClient";
import "./api/fetchInterceptors";
import "@xterm/xterm/css/xterm.css";
import "font-logos/assets/font-logos.css";
import "./index.css";

void loadConfig().then(() => {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <ErrorBoundary>
          <ClipboardProvider>
            <BrowserRouter basename="/">
              <App />
            </BrowserRouter>
          </ClipboardProvider>
        </ErrorBoundary>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </React.StrictMode>,
  );
});
