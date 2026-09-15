import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import CreateUserDrawer from "../CreateUserDrawer";

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

const Wrapper = createTestWrapper();
const createSpy = vi.fn();

function renderDrawer(
  overrides: Partial<{ open: boolean; onClose: () => void }> = {},
) {
  const defaults = { open: true, onClose: vi.fn() };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    ...render(<CreateUserDrawer {...props} />, { wrapper: Wrapper }),
  };
}

async function fillForm({
  name = "Alice",
  username = "alice",
  email = "alice@example.com",
  password = "pass123",
}: Partial<{
  name: string;
  username: string;
  email: string;
  password: string;
}> = {}) {
  if (name) await userEvent.type(screen.getByLabelText(/^name$/i), name);
  if (username)
    await userEvent.type(screen.getByLabelText(/^username$/i), username);
  if (email) await userEvent.type(screen.getByLabelText(/^email$/i), email);
  if (password)
    await userEvent.type(screen.getByLabelText(/^password$/i), password);
}

function submitButton() {
  return screen.getByRole("button", { name: /create user/i });
}

function setCreateError(status: number) {
  server.use(
    http.post("*/admin/api/users", () => HttpResponse.json({}, { status })),
  );
}

describe("CreateUserDrawer", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    createSpy.mockReset();
    server.use(
      http.post("*/admin/api/users", async ({ request }) => {
        createSpy({ body: await request.json() });
        return HttpResponse.json({});
      }),
    );
  });

  describe("form enabling", () => {
    it("enables submit when all required fields are filled", async () => {
      renderDrawer();
      await fillForm();
      expect(submitButton()).not.toBeDisabled();
    });

    it.each(["name", "username", "email", "password"] as const)(
      "keeps submit disabled when %s is missing",
      async (field) => {
        renderDrawer();
        await fillForm({ [field]: "" });
        expect(submitButton()).toBeDisabled();
      },
    );
  });

  describe("namespace limit controls", () => {
    it("shows the disable toggle and the max namespaces input once the limit is enabled", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      expect(
        screen.getByLabelText(/disable namespace creation/i),
      ).toBeInTheDocument();
      expect(screen.getByLabelText(/max namespaces/i)).toBeInTheDocument();
    });

    it("hides max namespaces input when 'Disable namespace creation' is checked", async () => {
      renderDrawer();
      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      await userEvent.click(
        screen.getByLabelText(/disable namespace creation/i),
      );
      expect(
        screen.queryByLabelText(/max namespaces/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("submit — success", () => {
    it("posts the form values and closes the drawer", async () => {
      const { onClose } = renderDrawer();
      await fillForm();

      await userEvent.click(submitButton());

      await waitFor(() => {
        expect(createSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({
              name: "Alice",
              username: "alice",
              email: "alice@example.com",
              password: "pass123",
              admin: false,
            }),
          }),
        );
      });
      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });

    it("omits max_namespaces when the limit is not enabled", async () => {
      renderDrawer();
      await fillForm();

      await userEvent.click(submitButton());

      await waitFor(() => {
        expect(createSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.not.objectContaining({
              max_namespaces: expect.anything(),
            }),
          }),
        );
      });
    });

    it("sends max_namespaces as 0 when namespace creation is disabled", async () => {
      renderDrawer();
      await fillForm();

      await userEvent.click(
        screen.getByLabelText(/set namespace creation limit/i),
      );
      await userEvent.click(
        screen.getByLabelText(/disable namespace creation/i),
      );
      await userEvent.click(submitButton());

      await waitFor(() => {
        expect(createSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ max_namespaces: 0 }),
          }),
        );
      });
    });
  });

  describe("submit — error handling", () => {
    it.each([
      [409, /already exists/i],
      [400, /failed to create user/i],
    ])("a %i response reports '%s'", async (status, message) => {
      setCreateError(status);
      renderDrawer();
      await fillForm();

      await userEvent.click(submitButton());

      await waitFor(() =>
        expect(screen.getByText(message)).toBeInTheDocument(),
      );
    });

    it("shows generic error for unexpected failures", async () => {
      server.use(http.post("*/admin/api/users", () => HttpResponse.error()));
      renderDrawer();
      await fillForm();

      await userEvent.click(submitButton());

      await waitFor(() =>
        expect(screen.getByText(/failed to create user/i)).toBeInTheDocument(),
      );
    });
  });

  describe("client-side validation", () => {
    it.each([
      [/^username$/i, "Alice", null],
      [/^email$/i, "not-an-email", /enter a valid email address/i],
      [/^password$/i, "abc", /5–32 characters/i],
    ] as const)(
      "marks %s invalid on blur and reports %s",
      async (label, value, message) => {
        renderDrawer();
        const input = screen.getByLabelText(label);
        await userEvent.type(input, value);
        await userEvent.tab();
        expect(input).toHaveAttribute("aria-invalid", "true");
        if (message)
          expect(await screen.findByText(message)).toBeInTheDocument();
      },
    );

    it("blocks submit when fields are non-empty but format is invalid", async () => {
      renderDrawer();
      await fillForm({ username: "Alice", email: "bad", password: "abc" });

      const submit = submitButton();
      expect(submit).toBeDisabled();
      await userEvent.click(submit);

      expect(createSpy).not.toHaveBeenCalled();
      expect(screen.getByLabelText(/^username$/i)).toHaveAttribute(
        "aria-invalid",
        "true",
      );
      expect(screen.getByLabelText(/^email$/i)).toHaveAttribute(
        "aria-invalid",
        "true",
      );
      expect(screen.getByLabelText(/^password$/i)).toHaveAttribute(
        "aria-invalid",
        "true",
      );
    });

    it("clears the invalid state when the user edits the field to a valid value", async () => {
      renderDrawer();
      const usernameInput = screen.getByLabelText(/^username$/i);
      await userEvent.type(usernameInput, "Alice");
      await userEvent.tab();
      expect(usernameInput).toHaveAttribute("aria-invalid", "true");

      await userEvent.clear(usernameInput);
      await userEvent.type(usernameInput, "alice");
      expect(usernameInput).not.toHaveAttribute("aria-invalid");
    });
  });

  describe("state reset on reopen", () => {
    it("clears the name field when closed then reopened", async () => {
      const { rerender } = renderDrawer();
      await userEvent.type(screen.getByLabelText(/^name$/i), "Alice");

      rerender(<CreateUserDrawer open={false} onClose={vi.fn()} />);
      rerender(<CreateUserDrawer open={true} onClose={vi.fn()} />);

      expect(screen.getByLabelText(/^name$/i)).toHaveValue("");
    });

    it("clears any error when closed then reopened", async () => {
      server.use(http.post("*/admin/api/users", () => HttpResponse.error()));
      const { rerender } = renderDrawer();
      await fillForm();
      await userEvent.click(submitButton());
      await waitFor(() => screen.getByRole("alert"));

      rerender(<CreateUserDrawer open={false} onClose={vi.fn()} />);
      rerender(<CreateUserDrawer open={true} onClose={vi.fn()} />);

      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });
});
