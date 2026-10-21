import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockTelemetryStatus } from "@/tests/factories";
import type { GetTelemetryResponse } from "@/client";
import AdminTelemetry from "../Telemetry";

const configureSpy = vi.fn();

function setStatus(overrides: Partial<GetTelemetryResponse> = {}) {
  const status = mockTelemetryStatus(overrides);
  server.use(http.get("*/admin/api/telemetry", () => HttpResponse.json(status)));
  return status;
}

async function renderPage() {
  render(<AdminTelemetry />, { wrapper: createTestWrapper() });
  return screen.findByRole("switch", { name: "Toggle telemetry" });
}

beforeEach(() => {
  configureSpy.mockReset();
  server.use(
    http.put("*/admin/api/telemetry", async ({ request }) => {
      configureSpy(await request.json());
      return HttpResponse.json({});
    }),
  );
});

describe("AdminTelemetry", () => {
  it("shows the administrator's setting on the switch", async () => {
    setStatus({ enabled: true });

    const toggle = await renderPage();

    expect(toggle).toBeChecked();
    expect(toggle).toBeEnabled();
  });

  it("turns telemetry off with the flipped value", async () => {
    const user = userEvent.setup();
    setStatus({ enabled: true });

    const toggle = await renderPage();
    await user.click(toggle);

    await waitFor(() => expect(configureSpy).toHaveBeenCalledWith({ enabled: false }));
  });

  it("locks the switch off when the environment disables telemetry", async () => {
    setStatus({ enabled: true, disabled_by_environment: true });

    const toggle = await renderPage();

    expect(toggle).not.toBeChecked();
    expect(toggle).toBeDisabled();
    expect(screen.getByText(/disabled by environment/i)).toBeInTheDocument();
  });

  it("previews the exact events the next heartbeat sends", async () => {
    const { preview } = setStatus();

    await renderPage();

    const region = screen.getByRole("region", { name: "Next heartbeat" });
    const payload: unknown = JSON.parse(within(region).getByLabelText("Heartbeat payload").textContent ?? "");
    expect(payload).toEqual(preview);
  });

  it("says when the settings cannot be loaded", async () => {
    server.use(
      http.get("*/admin/api/telemetry", () =>
        HttpResponse.json({ message: "boom" }, { status: 500 }),
      ),
    );

    render(<AdminTelemetry />, { wrapper: createTestWrapper() });

    expect(await screen.findByText(/failed to load telemetry settings/i)).toBeInTheDocument();
  });
});
