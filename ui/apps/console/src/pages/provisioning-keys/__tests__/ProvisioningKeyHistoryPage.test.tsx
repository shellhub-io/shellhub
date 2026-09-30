import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import {
  mockProvisioningKey,
  mockProvisioningKeyEvent,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { fullText } from "@/tests/fullText";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import ProvisioningKeyHistoryPage from "../ProvisioningKeyHistoryPage";

let reveals = 0;

function renderPage(
  key = mockProvisioningKey({
    id: "fleet-digest",
    name: "fleet-key",
    key_hint: "7e015c8d",
    usage_limit: 10,
    used_times: 3,
  }),
) {
  server.use(
    http.get("*/api/namespaces/provisioning-key", () => jsonWithTotal([key])),
    http.get("*/api/namespaces/provisioning-key/:key/reveal", () => {
      reveals += 1;
      return HttpResponse.json({ key: "the-secret" });
    }),
  );
  return render(
    <ClipboardProvider>
      <Routes>
        <Route
          path="/devices/add/fleet/:id/activity"
          element={<ProvisioningKeyHistoryPage />}
        />
      </Routes>
    </ClipboardProvider>,
    {
      wrapper: createTestWrapper({
        initialEntries: [`/devices/add/fleet/${key.id}/activity`],
      }),
    },
  );
}

beforeEach(() => {
  reveals = 0;
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces/provisioning-key/:id/history", () =>
      jsonWithTotal([
        mockProvisioningKeyEvent({
          id: "waiting",
          device_uid: "device-2",
          hostname: "kiosk-02",
          device_status: "pending",
          timestamp: "2024-03-05T10:00:00Z",
        }),
        mockProvisioningKeyEvent({
          id: "accepted",
          hostname: "edge-01",
          decided_status: "accepted",
          timestamp: "2024-03-04T09:00:00Z",
        }),
      ]),
    ),
  );
});

describe("Provisioning key page", () => {
  it("counts the key's uses against its allowance", async () => {
    renderPage();

    expect(
      await screen.findByText(fullText("3 of 10 devices")),
    ).toBeInTheDocument();
  });

  it("offers a decision on a device still waiting", async () => {
    renderPage();

    expect(await screen.findByText("Waiting for you")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Accept" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Reject" })).toBeInTheDocument();
  });

  it("offers a decision again on a rejected device sent back to pending", async () => {
    server.use(
      http.get("*/api/namespaces/provisioning-key/:id/history", () =>
        jsonWithTotal([
          mockProvisioningKeyEvent({
            hostname: "kiosk-03",
            decided_status: "rejected",
            device_status: "pending",
          }),
        ]),
      ),
    );
    renderPage();

    expect(await screen.findByText("Waiting for you")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Reject" })).toBeInTheDocument();
    expect(screen.queryByText(/Accept instead/)).not.toBeInTheDocument();
  });

  it("groups registrations by the day they happened", async () => {
    renderPage();

    expect(await screen.findByText("Mar 5, 2024")).toBeInTheDocument();
    expect(screen.getByText("Mar 4, 2024")).toBeInTheDocument();
  });

  it("fetches the key's secret only when asked to show it", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: /Details/ }));
    expect(screen.getByText(/7e015c8d•+/)).toBeInTheDocument();
    expect(reveals).toBe(0);

    await user.click(screen.getByRole("button", { name: "Show" }));

    expect(await screen.findByText("the-secret")).toBeInTheDocument();
    expect(reveals).toBe(1);
  });

  it("opens the details of a webhook key without being asked", async () => {
    renderPage(
      mockProvisioningKey({
        id: "hook-digest",
        name: "hook-key",
        mode: "webhook",
        webhook_url: "https://hooks.example.com/v1/enroll/devices",
      }),
    );

    expect(
      await screen.findByText("https://hooks.example.com/v1/enroll/devices"),
    ).toBeInTheDocument();
  });
});
