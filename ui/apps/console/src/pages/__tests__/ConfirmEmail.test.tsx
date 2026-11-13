import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import ConfirmEmail from "../ConfirmEmail";

function renderWithParams(search = "?id=uid123&token=tok456") {
  return render(<ConfirmEmail />, {
    wrapper: createTestWrapper({ initialEntries: [`/confirm-email${search}`] }),
  });
}

function answer(status: number) {
  const sent: { uid?: string; token?: string }[] = [];
  server.use(
    http.post("*/api/user/:uid/confirm_email", async ({ params, request }) => {
      const body = (await request.json()) as { token: string };
      sent.push({ uid: String(params.uid), token: body.token });
      return status === 200
        ? new HttpResponse(null, { status })
        : HttpResponse.json({ message: "error" }, { status });
    }),
  );
  return sent;
}

describe("ConfirmEmail", () => {
  it.each(["?id=&token=", "?token=tok456", "?id=uid123"])(
    "offers no confirmation for an incomplete link %s",
    (search) => {
      renderWithParams(search);

      expect(
        screen.getByText(/invalid confirmation link/i),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /confirm email/i }),
      ).not.toBeInTheDocument();
    },
  );

  it("spends nothing until the button is pressed", () => {
    const queryClient = createTestQueryClient();
    render(<ConfirmEmail />, {
      wrapper: createTestWrapper({
        queryClient,
        initialEntries: ["/confirm-email?id=uid123&token=tok456"],
      }),
    });

    expect(
      screen.getByRole("button", { name: /confirm email/i }),
    ).toBeEnabled();
    expect(queryClient.getMutationCache().getAll()).toEqual([]);
  });

  it("sends the link's user and token and reports the email confirmed", async () => {
    const sent = answer(200);
    const user = userEvent.setup();
    renderWithParams();

    await user.click(screen.getByRole("button", { name: /confirm email/i }));

    expect(await screen.findByText(/email confirmed/i)).toBeInTheDocument();
    expect(sent).toEqual([{ uid: "uid123", token: "tok456" }]);
  });

  it.each([
    [401, /invalid, expired or already used/i],
    [409, /another account took this address/i],
    [400, /can no longer be used/i],
  ])("explains a %s answer", async (status, message) => {
    answer(status);
    const user = userEvent.setup();
    renderWithParams();

    await user.click(screen.getByRole("button", { name: /confirm email/i }));

    expect(await screen.findByText(message)).toBeInTheDocument();
  });
});
