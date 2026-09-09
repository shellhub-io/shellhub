import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockUserAuth } from "@/tests/factories";
import { useNamespaceCreateForm } from "@/hooks/useNamespaceCreateForm";

const createSpy = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  server.use(
    http.post("*/api/namespaces", async ({ request }) => {
      createSpy(await request.json());
      return HttpResponse.json(mockNamespace({ name: "my-ns" }));
    }),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json(mockUserAuth({ token: "jwt-token" })),
    ),
  );
});

function Harness({ onCreated }: { onCreated?: () => void }) {
  const form = useNamespaceCreateForm(onCreated);

  return (
    <form onSubmit={(e) => void form.submit(e)}>
      <input
        aria-label="Namespace Name"
        value={form.name}
        onChange={(e) => form.changeName(e.target.value)}
      />
      {form.error && <p role="alert">{form.error}</p>}
      <button type="submit">Create</button>
    </form>
  );
}

function renderForm(onCreated?: () => void) {
  render(<Harness onCreated={onCreated} />, { wrapper: createTestWrapper() });
  return {
    nameInput: () => screen.getByLabelText("Namespace Name"),
    create: () => screen.getByRole("button", { name: "Create" }),
  };
}

async function submitName(name: string) {
  const user = userEvent.setup();
  const form = renderForm();
  await user.type(form.nameInput(), name);
  await user.click(form.create());
  return user;
}

describe("useNamespaceCreateForm", () => {
  it.each([
    [409, "A namespace with this name already exists."],
    [403, "You have reached the namespace limit or do not have permission."],
    [400, "The namespace name is invalid."],
    [500, "An unexpected error occurred. Please try again."],
  ])("reports %s as '%s'", async (status, message) => {
    server.use(
      http.post("*/api/namespaces", () => HttpResponse.json({}, { status })),
    );

    await submitName("my-ns");

    expect(await screen.findByRole("alert")).toHaveTextContent(message);
  });

  it("reports a network failure with the generic message", async () => {
    server.use(http.post("*/api/namespaces", () => HttpResponse.error()));

    await submitName("my-ns");

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "An unexpected error occurred. Please try again.",
    );
  });

  it("rejects an invalid name without sending a request", async () => {
    await submitName("ab");

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Name must be at least 3 characters",
    );
    expect(createSpy).not.toHaveBeenCalled();
  });

  it("clears the error once the name changes again", async () => {
    server.use(
      http.post("*/api/namespaces", () => HttpResponse.json({}, { status: 409 })),
    );

    const user = await submitName("my-ns");
    expect(await screen.findByRole("alert")).toBeInTheDocument();

    await user.type(screen.getByLabelText("Namespace Name"), "x");

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("sends the typed name and runs onCreated once the namespace exists", async () => {
    const onCreated = vi.fn();
    const user = userEvent.setup();
    const form = renderForm(onCreated);

    await user.type(form.nameInput(), "my-ns");
    await user.click(form.create());

    await waitFor(() => expect(onCreated).toHaveBeenCalledOnce());
    expect(createSpy).toHaveBeenCalledWith({ name: "my-ns" });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
