import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import type { DeviceStatus } from "@/client";
import SegmentedTabs, { type SegmentedTab } from "../SegmentedTabs";

const STATUSES: SegmentedTab<DeviceStatus>[] = [
  { label: "Accepted", value: "accepted" },
  { label: "Pending", value: "pending" },
  { label: "Rejected", value: "rejected" },
];

function Tabs({
  initial = "accepted",
  tabs = STATUSES,
}: {
  initial?: DeviceStatus;
  tabs?: SegmentedTab<DeviceStatus>[];
}) {
  const [value, setValue] = useState<DeviceStatus>(initial);
  return (
    <SegmentedTabs tabs={tabs} value={value} onChange={setValue} label="Device status" />
  );
}

const tab = (status: DeviceStatus) =>
  screen.getByRole("tab", { name: new RegExp(`^${status}$`, "i") });

async function pressOn(start: DeviceStatus, key: string) {
  const user = userEvent.setup();
  await user.tab();
  expect(tab(start)).toHaveFocus();
  await user.keyboard(key);
}

describe("SegmentedTabs", () => {
  it("keeps only the selected tab in the tab order", () => {
    render(<Tabs initial="pending" />);

    expect(tab("accepted")).toHaveAttribute("tabindex", "-1");
    expect(tab("pending")).toHaveAttribute("tabindex", "0");
    expect(tab("rejected")).toHaveAttribute("tabindex", "-1");
  });

  it("selects a tab on click", async () => {
    const user = userEvent.setup();
    render(<Tabs />);

    await user.click(tab("rejected"));

    expect(tab("rejected")).toHaveAttribute("aria-selected", "true");
  });

  it.each([
    { start: "accepted", key: "{ArrowRight}", lands: "pending" },
    { start: "rejected", key: "{ArrowRight}", lands: "accepted" },
    { start: "pending", key: "{ArrowLeft}", lands: "accepted" },
    { start: "accepted", key: "{ArrowLeft}", lands: "rejected" },
    { start: "pending", key: "{Home}", lands: "accepted" },
    { start: "pending", key: "{End}", lands: "rejected" },
  ] as const)("moves from $start with $key to $lands, selecting and focusing it", async ({ start, key, lands }) => {
    render(<Tabs initial={start} />);

    await pressOn(start, key);

    expect(tab(lands)).toHaveAttribute("aria-selected", "true");
    expect(tab(lands)).toHaveFocus();
  });

  it("leaves the first tab focusable when the value matches no tab", () => {
    render(<Tabs tabs={STATUSES.slice(1)} />);

    expect(tab("pending")).toHaveAttribute("tabindex", "0");
  });
});
