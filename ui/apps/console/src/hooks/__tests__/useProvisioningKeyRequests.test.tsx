import { describe, it, expect } from "vitest";
import { waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHookWithClient } from "@/tests/wrapper";
import { useRevealProvisioningKey } from "../useProvisioningKeyRequests";

const REVEAL = "*/api/namespaces/provisioning-key/:key/reveal";

function countReveals() {
  const revealed: string[] = [];
  server.use(
    http.get(REVEAL, ({ params }) => {
      revealed.push(String(params.key));
      return HttpResponse.json({ key: "plaintext" });
    }),
  );
  return revealed;
}

describe("useRevealProvisioningKey", () => {
  it.each([
    ["no key is named", null, true],
    ["the caller has not opted in", "fleet-key", false],
  ])("decrypts nothing when %s", (_, name, enabled) => {
    const { result } = renderHookWithClient(() =>
      useRevealProvisioningKey(name, enabled),
    );

    expect(result.current.isLoading).toBe(false);
    expect(result.current.key).toBe("");
  });

  it("decrypts the named key once the caller opts in", async () => {
    const revealed = countReveals();
    const { result } = renderHookWithClient(() =>
      useRevealProvisioningKey("fleet-key", true),
    );

    await waitFor(() => expect(result.current.key).toBe("plaintext"));
    expect(revealed).toEqual(["fleet-key"]);
  });
});
