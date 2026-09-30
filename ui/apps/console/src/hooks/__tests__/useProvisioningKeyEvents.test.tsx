import { describe, it, expect } from "vitest";
import { act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { mockProvisioningKeyEvent } from "@/tests/factories";
import { renderHookWithClient } from "@/tests/wrapper";
import { useProvisioningKeyEvents } from "../useProvisioningKeyEvents";

const HISTORY = "*/api/namespaces/provisioning-key/:id/history";

function events(...ids: string[]) {
  return ids.map((id) => mockProvisioningKeyEvent({ id }));
}

describe("useProvisioningKeyEvents", () => {
  it("keeps an event once when a new registration shifts it onto the next page", async () => {
    server.use(
      http.get(HISTORY, ({ request }) => {
        const page = new URL(request.url).searchParams.get("page");
        return page === "1"
          ? jsonWithTotal(events("c", "b"), 4)
          : jsonWithTotal(events("b", "a"), 4);
      }),
    );
    const { result } = renderHookWithClient(() =>
      useProvisioningKeyEvents({ id: "key-digest-1", perPage: 2 }),
    );
    await waitFor(() => expect(result.current.events).toHaveLength(2));

    act(() => result.current.loadMore());

    await waitFor(() =>
      expect(result.current.events.map((e) => e.id)).toEqual(["c", "b", "a"]),
    );
  });

  it("tells a failed later page apart from a failed first load", async () => {
    server.use(
      http.get(HISTORY, ({ request }) =>
        new URL(request.url).searchParams.get("page") === "1"
          ? jsonWithTotal(events("b", "a"), 4)
          : new HttpResponse(null, { status: 500 }),
      ),
    );
    const { result } = renderHookWithClient(() =>
      useProvisioningKeyEvents({ id: "key-digest-1", perPage: 2 }),
    );
    await waitFor(() => expect(result.current.events).toHaveLength(2));

    act(() => result.current.loadMore());

    await waitFor(() => expect(result.current.loadMoreFailed).toBe(true));
    expect(result.current.events.map((e) => e.id)).toEqual(["b", "a"]);
  });
});
