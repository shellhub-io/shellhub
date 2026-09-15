import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockUserAuth } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import CreateNamespace from "../CreateNamespace";

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.post("*/api/namespaces", () =>
      HttpResponse.json(mockNamespace({ name: "my-ns" })),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json(mockUserAuth({ token: "jwt-token" })),
    ),
  );
});

function renderComponent() {
  return render(<CreateNamespace />, { wrapper: createTestWrapper() });
}

describe("CreateNamespace — CloudForm", () => {
  it("shows 'A namespace with this name already exists.' on 409", async () => {
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 409 }),
      ),
    );
    const user = userEvent.setup();
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();
  });

  it("shows the limit/permission message on 403", async () => {
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    const user = userEvent.setup();
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText(
        "You have reached the namespace limit or do not have permission.",
      ),
    ).toBeInTheDocument();
  });

  it("shows the invalid-name message on 400", async () => {
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 400 }),
      ),
    );
    const user = userEvent.setup();
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("The namespace name is invalid."),
    ).toBeInTheDocument();
  });

  it("shows the generic fallback message on 500", async () => {
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText(
        "An unexpected error occurred. Please try again.",
      ),
    ).toBeInTheDocument();
  });

  it("clears the error text when the user types after a failed submission", async () => {
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 409 }),
      ),
    );
    const user = userEvent.setup();
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();

    await user.type(screen.getByPlaceholderText("my-namespace"), "x");
    expect(
      screen.queryByText("A namespace with this name already exists."),
    ).not.toBeInTheDocument();
  });
});
