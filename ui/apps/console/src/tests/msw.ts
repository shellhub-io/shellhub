import { setupServer } from "msw/node";
import { http, HttpResponse } from "msw";
import { mockTag } from "./factories";

/** MSW server shared by every test file — started once in setup.ts. */
export const server = setupServer();

/** JSON response with `X-Total-Count` for paginated endpoints. */
export function jsonWithTotal<T>(data: T[], total = data.length) {
  return HttpResponse.json(data, {
    headers: { "X-Total-Count": String(total) },
  });
}

/** Override the tags endpoint with the given tag names. */
export function setTags(names: string[]) {
  server.use(
    http.get("*/api/tags", () =>
      jsonWithTotal(names.map((name) => mockTag({ name }))),
    ),
  );
}
