import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice } from "@/tests/factories";
import DepartingDevices from "../DepartingDevices";

function serveDevices(count: number) {
  const served = { requests: 0 };
  server.use(
    http.get("*/api/devices", () => {
      served.requests++;
      return jsonWithTotal(
        Array.from({ length: count }, (_, i) =>
          mockDevice({ uid: `uid-${i}`, name: `device-${i}` }),
        ),
      );
    }),
  );
  return served;
}

describe("DepartingDevices", () => {
  it("tells a leaving member their devices go with them", async () => {
    serveDevices(1);
    render(<DepartingDevices memberId="user-7" self />, {
      wrapper: createTestWrapper(),
    });

    expect(
      await screen.findByText(
        "1 device you paired will be removed with you and go back to pairing.",
      ),
    ).toBeInTheDocument();
  });

  it("lists the first few devices and counts the rest", async () => {
    serveDevices(7);
    render(<DepartingDevices memberId="user-7" />, {
      wrapper: createTestWrapper(),
    });

    const region = await screen.findByRole("region", {
      name: "Paired devices",
    });
    expect(screen.getAllByRole("listitem")).toHaveLength(5);
    expect(region).toHaveTextContent("and 2 more");
  });

  it("renders nothing for a member who paired no device", async () => {
    const served = serveDevices(0);
    const { container } = render(<DepartingDevices memberId="user-7" />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => expect(served.requests).toBe(1));
    expect(container).toBeEmptyDOMElement();
  });
});
