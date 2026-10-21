import { describe, it, expect } from "vitest";
import { act } from "@testing-library/react";
import { useMutation } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHookWithClient } from "@/tests/wrapper";
import { deleteFirewallRuleMutation, type FirewallRulesResponse } from "@/client";
import { mockFirewallRule } from "@/tests/factories";
import { useRowAction } from "../useRowAction";

const RULE = mockFirewallRule();
const OTHER_RULE = mockFirewallRule({ id: "rule-2" });

function respondWith(status: number) {
  server.use(
    http.delete(
      "*/api/firewall/rules/:id",
      () => new HttpResponse(status === 200 ? null : "{}", { status }),
    ),
  );
}

function renderRowAction(verb?: string) {
  return renderHookWithClient(() => {
    const deleteRule = useMutation(deleteFirewallRuleMutation());
    return useRowAction({
      item: "firewall rule",
      verb,
      run: (rule: FirewallRulesResponse) => deleteRule.mutateAsync({ path: { id: rule.id } }),
    });
  });
}

async function openAndConfirm(result: ReturnType<typeof renderRowAction>["result"]) {
  act(() => result.current.open(RULE));
  await act(() => result.current.confirm());
}

describe("useRowAction", () => {
  it.each([
    [undefined, 403, "Failed to delete firewall rule. You do not have permission to do this."],
    ["revoke", 404, "Failed to revoke firewall rule. We could not find what you asked for."],
    [undefined, 418, "Failed to delete firewall rule. Something went wrong. Please try again."],
  ])("names the verb %s and the %i status in the error", async (verb, status, message) => {
    respondWith(status);
    const { result } = renderRowAction(verb);

    await openAndConfirm(result);

    expect(result.current.error).toBe(message);
    expect(result.current.target).toBe(RULE);
  });

  it("clears the error on close", async () => {
    respondWith(500);
    const { result } = renderRowAction();
    await openAndConfirm(result);

    act(() => result.current.close());

    expect(result.current.error).toBeNull();
    expect(result.current.target).toBeNull();
  });

  it("clears the error when another row is opened", async () => {
    respondWith(500);
    const { result } = renderRowAction();
    await openAndConfirm(result);

    act(() => result.current.open(OTHER_RULE));

    expect(result.current.error).toBeNull();
    expect(result.current.target).toBe(OTHER_RULE);
  });

  it("closes after the action succeeds", async () => {
    respondWith(200);
    const { result } = renderRowAction();

    await openAndConfirm(result);

    expect(result.current.target).toBeNull();
    expect(result.current.error).toBeNull();
  });
});
