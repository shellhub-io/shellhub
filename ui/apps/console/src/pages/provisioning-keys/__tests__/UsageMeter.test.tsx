import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { mockProvisioningKey } from "@/tests/factories";
import { fullText } from "@/tests/fullText";
import UsageMeter from "../UsageMeter";

function renderMeter(overrides: Parameters<typeof mockProvisioningKey>[0]) {
  render(<UsageMeter provisioningKey={mockProvisioningKey(overrides)} />);
}

describe("UsageMeter", () => {
  it.each([
    {
      name: "a spent key says its limit is reached",
      key: { usage_limit: 5, used_times: 5 },
      count: "5 of 5 devices",
      detail: "limit reached",
    },
    {
      name: "devices waiting on a decision are counted",
      key: { usage_limit: 10, used_times: 3, pending_devices: 2 },
      count: "3 of 10 devices",
      detail: "2 waiting",
    },
    {
      name: "waiting devices past the limit are called out",
      key: { usage_limit: 4, used_times: 3, pending_devices: 3 },
      count: "3 of 4 devices",
      detail: "3 waiting · 2 over",
    },
  ])("$name", ({ key, count, detail }) => {
    renderMeter(key);

    expect(screen.getByText(fullText(count))).toBeInTheDocument();
    expect(screen.getByText(detail)).toBeInTheDocument();
  });

  it.each([
    {
      name: "an unlimited key",
      key: { usage_limit: 0, used_times: 4 },
      count: "4 of unlimited devices",
    },
    {
      name: "a limited key with room left",
      key: { usage_limit: 10, used_times: 3 },
      count: "3 of 10 devices",
    },
  ])("counts $name on one line", ({ key, count }) => {
    renderMeter(key);

    expect(screen.getByText(fullText(count))).toBeInTheDocument();
    expect(
      screen.queryByText(/waiting|limit reached|enrolling/),
    ).not.toBeInTheDocument();
  });

  it.each([
    { name: "revoked", key: { revoked: true } },
    { name: "disabled", key: { disabled: true } },
    { name: "expired", key: { expires_at: "2020-01-01T00:00:00Z" } },
  ])("claims no uses left on a $name key", ({ key }) => {
    renderMeter({ usage_limit: 10, used_times: 3, ...key });

    expect(screen.getByText("not enrolling")).toBeInTheDocument();
  });

  it("names a spent key's state for a pointer", () => {
    renderMeter({ usage_limit: 5, used_times: 5 });

    expect(screen.getByTitle("Limit reached")).toBeInTheDocument();
  });
});
