import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import CreateNamespace from "../CreateNamespace";

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.post("*/api/namespaces", () => HttpResponse.json({}, { status: 409 })),
  );
});

describe("CreateNamespace — NamespaceCreateForm", () => {
  it("surfaces a failed creation on the name field", async () => {
    const user = userEvent.setup();
    render(<CreateNamespace />, { wrapper: createTestWrapper() });

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();
  });
});
