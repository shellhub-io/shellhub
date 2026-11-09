import { beforeEach, describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockNamespace } from "@/tests/factories";
import SshidCard from "../SshidCard";

function renderCard() {
  render(
    <SshidCard uid="abc" name="zurich-1" />,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  seedAuthStore();
});

describe("SshidCard", () => {
  it("shows the SSHID in the active namespace", async () => {
    server.use(
      http.get("*/api/namespaces/:tenant", () =>
        HttpResponse.json(mockNamespace({ name: "dev" })),
      ),
    );

    renderCard();

    expect(
      await screen.findByText("dev.zurich-1@localhost"),
    ).toBeInTheDocument();
  });

  it("shows the uid until the namespace has loaded", () => {
    server.use(
      http.get("*/api/namespaces/:tenant", () => new Promise(() => {})),
    );

    renderCard();

    expect(screen.getByText("abc")).toBeInTheDocument();
  });
});
