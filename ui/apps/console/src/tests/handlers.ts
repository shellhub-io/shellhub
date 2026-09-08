import { http, HttpResponse } from "msw";
import { jsonWithTotal } from "./msw";
import { mockNamespace, mockStats } from "./factories";

/** Handlers for the ambient queries most components fire on mount. */
export const defaultHandlers = [
  http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
  http.get("*/api/namespaces/:tenant", () =>
    HttpResponse.json(mockNamespace()),
  ),
  http.get("*/api/stats", () => HttpResponse.json(mockStats())),
  http.get("*/api/tags", () => jsonWithTotal([])),
];
