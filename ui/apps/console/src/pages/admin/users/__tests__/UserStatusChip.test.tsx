import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { mockAdminUser } from "@/tests/factories";
import UserStatusChip from "../UserStatusChip";
import { userStatus } from "../userStatus";

describe("UserStatusChip", () => {
  it.each([
    ["confirmed", "Confirmed"],
    ["not-confirmed", "Not Confirmed"],
  ] as const)("renders '%s' status as '%s'", (status, label) => {
    render(<UserStatusChip status={status} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });
});

describe("userStatus", () => {
  it("reads a user awaiting approval as such, whatever the confirmation says", () => {
    expect(
      userStatus(
        mockAdminUser({ status: "confirmed", awaiting_approval: true }),
      ),
    ).toBe("awaiting_approval");
  });

  it.each(["confirmed", "not-confirmed"] as const)(
    "reads a user not awaiting approval by the confirmation status, %s",
    (status) => {
      expect(
        userStatus(mockAdminUser({ status, awaiting_approval: false })),
      ).toBe(status);
    },
  );
});
