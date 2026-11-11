import { http, HttpResponse } from "msw";
import { jsonWithTotal } from "./msw";
import { mockDevice, mockNamespace, mockStats } from "./factories";

/** Handlers for the ambient queries most components fire on mount. */
export const defaultHandlers = [
  http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
  http.get("*/api/namespaces/:tenant", () =>
    HttpResponse.json(mockNamespace()),
  ),
  http.get("*/api/stats", () => HttpResponse.json(mockStats())),
  http.get("*/api/tags", () => jsonWithTotal([])),
];

/**
 * Handlers for what the connect dialog fetches once it opens. The device handler answers any single
 * path segment under `/api/devices` with `mockDevice` defaults, so it also shadows an earlier
 * `/api/devices/tenant-registrations` or `/api/devices/resolve` handler, and the identities list is
 * empty. A test that asserts a device's name or a registered identity puts its own handler ahead
 * of these in the same `server.use` call, or in a later call: spreading these again in a later call
 * moves them back in front.
 */
export const connectDialogHandlers = [
  http.get("*/api/devices/:uid", ({ params }) =>
    HttpResponse.json(mockDevice({ uid: String(params.uid) })),
  ),
  http.get("*/api/ssh-identities", () => HttpResponse.json([])),
];
