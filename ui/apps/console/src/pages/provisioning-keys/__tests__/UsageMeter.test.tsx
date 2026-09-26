import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { mockProvisioningKey } from "@/tests/factories";
import UsageMeter from "../UsageMeter";

function renderMeter(overrides: Parameters<typeof mockProvisioningKey>[0]) {
  render(<UsageMeter provisioningKey={mockProvisioningKey(overrides)} />);
}

describe("UsageMeter", () => {
  it.each([
    {
      name: "an unlimited key counts its uses",
      key: { usage_limit: 0, used_times: 4 },
      count: "4 used",
      detail: "no limit",
    },
    {
      name: "a limited key says how many uses are left",
      key: { usage_limit: 10, used_times: 3 },
      count: "3 / 10",
      detail: "7 left",
    },
    {
      name: "a spent key says its limit is reached",
      key: { usage_limit: 5, used_times: 5 },
      count: "5 / 5",
      detail: "limit reached",
    },
    {
      name: "devices waiting on a decision are counted",
      key: { usage_limit: 10, used_times: 3, pending_devices: 2 },
      count: "3 / 10",
      detail: "2 waiting",
    },
    {
      name: "waiting devices past the limit are called out",
      key: { usage_limit: 4, used_times: 3, pending_devices: 3 },
      count: "3 / 4",
      detail: "3 waiting · 2 over",
    },
  ])("$name", ({ key, count, detail }) => {
    renderMeter(key);

    expect(screen.getByText(count)).toBeInTheDocument();
    expect(screen.getByText(detail)).toBeInTheDocument();
  });

  it.each([
    { name: "revoked", key: { revoked: true } },
    { name: "disabled", key: { disabled: true } },
    { name: "expired", key: { expires_at: "2020-01-01T00:00:00Z" } },
  ])("claims no uses left on a $name key", ({ key }) => {
    renderMeter({ usage_limit: 10, used_times: 3, ...key });

    expect(screen.getByText("not enrolling")).toBeInTheDocument();
    expect(screen.queryByText("7 left")).not.toBeInTheDocument();
  });

  it("names a spent key's state for a pointer", () => {
    renderMeter({ usage_limit: 5, used_times: 5 });

    expect(screen.getByTitle("Limit reached")).toBeInTheDocument();
  });
});
