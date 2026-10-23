import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent, { type UserEvent } from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockWebhook, mockWebhookDelivery } from "@/tests/factories";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

import WebhookSettings from "../WebhookSettings";

function renderPage() {
  render(<WebhookSettings />, { wrapper: createTestWrapper() });
}

async function fillWebhook(user: UserEvent, url: string, events: string[]) {
  await user.click(await screen.findByRole("button", { name: "Add webhook" }));
  await user.type(screen.getByLabelText("URL"), url);
  for (const event of events) {
    await user.click(screen.getByLabelText(event));
  }
}

async function addWebhook(user: UserEvent, url: string, events: string[]) {
  await fillWebhook(user, url, events);
  await user.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Add webhook" }));
}

async function editWebhookEvent(user: UserEvent, event: string) {
  await user.click(await screen.findByRole("button", { name: "Edit" }));
  await user.click(screen.getByLabelText(event));
  await user.click(screen.getByRole("button", { name: "Save" }));
}

describe("WebhookSettings", () => {
  beforeEach(() => {
    seedAuthStore({ role: "owner" });
    server.use(http.get("*/api/namespaces/webhooks", () => jsonWithTotal([mockWebhook()])));
  });

  it("lists each webhook with its URL, events and state", async () => {
    renderPage();

    const card = await screen.findByRole("article", { name: "https://hooks.example.com/shellhub" });
    expect(within(card).getByText("device.accepted")).toBeInTheDocument();
    expect(within(card).getByText("Enabled")).toBeInTheDocument();
  });

  it("shows the secret once a webhook is added", async () => {
    let sent: unknown;
    server.use(
      http.post("*/api/namespaces/webhooks", async ({ request }) => {
        sent = await request.json();
        return HttpResponse.json({ ...mockWebhook(), secret: "JBSWY3DPEHPK3PXP" });
      }),
    );
    const user = userEvent.setup();
    renderPage();

    await addWebhook(user, "https://hooks.example.com/new", ["Device pending", "Device removed"]);

    expect(await screen.findByText("JBSWY3DPEHPK3PXP")).toBeInTheDocument();
    expect(sent).toEqual({ url: "https://hooks.example.com/new", events: ["device.pending", "device.removed"] });
  });

  it.each(["ftp://hooks.example.com", "http://:80"])("refuses %s as a webhook URL", async (url) => {
    const user = userEvent.setup();
    renderPage();

    await fillWebhook(user, url, ["Device accepted"]);

    expect(await screen.findByText("Enter an http or https URL.")).toBeInTheDocument();
    expect(within(screen.getByRole("dialog")).getByRole("button", { name: "Add webhook" })).toBeDisabled();
  });

  it("reveals the secret on request", async () => {
    server.use(
      http.get("*/api/namespaces/webhooks/:id/reveal", () => HttpResponse.json({ secret: "REVEALEDSECRET" })),
    );
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Show secret" }));

    expect(await screen.findByText("REVEALEDSECRET")).toBeInTheDocument();
  });

  it("saves an edit with the changed events", async () => {
    let sent: unknown;
    server.use(
      http.patch("*/api/namespaces/webhooks/:id", async ({ request }) => {
        sent = await request.json();
        return HttpResponse.json(mockWebhook({ events: ["device.accepted", "device.rejected"] }));
      }),
    );
    const user = userEvent.setup();
    renderPage();

    await editWebhookEvent(user, "Device rejected");

    await vi.waitFor(() =>
      expect(sent).toEqual({ url: "https://hooks.example.com/shellhub", events: ["device.accepted", "device.rejected"] }),
    );
  });

  it("says an add was refused for the limit or a permission", async () => {
    server.use(http.post("*/api/namespaces/webhooks", () => new HttpResponse(null, { status: 403 })));
    const user = userEvent.setup();
    renderPage();

    await addWebhook(user, "https://hooks.example.com/new", ["Device pending"]);

    expect(
      await screen.findByText(
        "Couldn't add the webhook: the namespace has the most it can hold, or you don't have permission.",
      ),
    ).toBeInTheDocument();
  });

  it("says a refused edit is a missing permission", async () => {
    server.use(http.patch("*/api/namespaces/webhooks/:id", () => new HttpResponse(null, { status: 403 })));
    const user = userEvent.setup();
    renderPage();

    await editWebhookEvent(user, "Device rejected");

    expect(await screen.findByText("You don't have permission to edit this webhook.")).toBeInTheDocument();
  });

  it("pauses a webhook", async () => {
    let sent: unknown;
    server.use(
      http.patch("*/api/namespaces/webhooks/:id", async ({ request }) => {
        sent = await request.json();
        return HttpResponse.json(mockWebhook({ enabled: false }));
      }),
    );
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Pause" }));

    await vi.waitFor(() => expect(sent).toEqual({ enabled: false }));
  });

  it("says so when pausing a webhook fails", async () => {
    server.use(
      http.patch("*/api/namespaces/webhooks/:id", () => new HttpResponse(null, { status: 500 })),
    );
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Pause" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Couldn't update the webhook. Try again.");
  });

  it("deletes a webhook only after confirmation", async () => {
    const deleted = vi.fn();
    server.use(
      http.delete("*/api/namespaces/webhooks/:id", ({ params }) => {
        deleted(params.id);
        return new HttpResponse(null, { status: 200 });
      }),
    );
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Delete" }));
    expect(deleted).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Delete webhook" }));

    await vi.waitFor(() => expect(deleted).toHaveBeenCalledWith(mockWebhook().id));
  });

  it("shows how the recent deliveries went", async () => {
    server.use(
      http.get("*/api/namespaces/webhooks/:id/deliveries", () => jsonWithTotal([mockWebhookDelivery()])),
    );
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Show deliveries" }));

    const deliveries = await screen.findByRole("list", { name: "Recent deliveries" });
    expect(within(deliveries).getByText("failed")).toBeInTheDocument();
    expect(within(deliveries).getByText("webhook returned status 500")).toBeInTheDocument();
    expect(within(deliveries).getByText(/12 attempts, last answered 500/)).toBeInTheDocument();
  });

  it("offers an administrator's controls to an administrator only", async () => {
    seedAuthStore({ role: "operator" });
    renderPage();

    await screen.findByRole("article", { name: "https://hooks.example.com/shellhub" });
    for (const name of ["Add webhook", "Edit", "Show secret", "Delete", "Pause"]) {
      expect(screen.queryByRole("button", { name })).not.toBeInTheDocument();
    }
  });
});
