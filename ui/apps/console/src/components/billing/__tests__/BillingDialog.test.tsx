import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

vi.mock("@/hooks/useFocusTrap", () => ({ useFocusTrap: vi.fn() }));

const mockIsSdkError = vi.fn();
vi.mock("@/api/errors", () => ({
  isSdkError: (err: unknown): boolean => mockIsSdkError(err) as boolean,
}));

const mockMutateAsync = vi.fn();
const mockRefetchSubscription = vi.fn();
const mockIsPending = { value: false };

vi.mock("@/hooks/useBilling", () => ({
  useCreateSubscription: () => ({
    mutateAsync: (...args: unknown[]): Promise<unknown> =>
      mockMutateAsync(...args) as Promise<unknown>,
    isPending: mockIsPending.value,
  }),
  useSubscription: () => ({ refetch: mockRefetchSubscription }),
}));

vi.mock("../BillingPayment", () => ({
  default: ({
    onHasDefault,
    onNoPaymentMethods,
  }: {
    onHasDefault: () => void;
    onNoPaymentMethods: () => void;
  }) =>
    React.createElement(
      "div",
      { "data-testid": "billing-payment" },
      React.createElement(
        "button",
        { onClick: onHasDefault, "data-testid": "trigger-has-default" },
        "Set default",
      ),
      React.createElement(
        "button",
        {
          onClick: onNoPaymentMethods,
          "data-testid": "trigger-no-payment-methods",
        },
        "Clear default",
      ),
    ),
}));

vi.mock("../BillingLetter", () => ({
  default: () =>
    React.createElement("div", { "data-testid": "billing-letter" }),
}));

vi.mock("../BillingCheckout", () => ({
  default: () =>
    React.createElement("div", { "data-testid": "billing-checkout" }),
}));

vi.mock("../BillingSuccessful", () => ({
  default: () =>
    React.createElement("div", { "data-testid": "billing-successful" }),
}));

vi.mock("@/components/common/BaseDialog", () => ({
  default: ({
    open,
    children,
  }: {
    open: boolean;
    onClose: () => void;
    canClose?: () => boolean;
    children: React.ReactNode;
  }) => {
    if (!open) return null;
    return React.createElement("div", { role: "dialog" }, children);
  },
}));

import BillingDialog from "../BillingDialog";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockIsPending.value = false;
  mockRefetchSubscription.mockResolvedValue({ data: { status: "active" } });
  mockIsSdkError.mockReturnValue(false);
});

afterEach(cleanup);

function renderDialog(onClose = vi.fn(), onSuccess = vi.fn()) {
  return {
    onClose,
    onSuccess,
    ...render(
      <BillingDialog open={true} onClose={onClose} onSuccess={onSuccess} />,
      { wrapper: createWrapper() },
    ),
  };
}

async function goToStep2(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: /next/i }));
}

async function goToStep3(user: ReturnType<typeof userEvent.setup>) {
  await goToStep2(user);
  await user.click(screen.getByTestId("trigger-has-default"));
  await user.click(screen.getByRole("button", { name: /^next$/i }));
}

describe("BillingDialog", () => {
  describe("Step 1 — Overview", () => {
    it("renders BillingLetter on step 1", () => {
      renderDialog();
      expect(screen.getByTestId("billing-letter")).toBeInTheDocument();
    });

    it("shows a 'Next' button on step 1", () => {
      renderDialog();
      expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
    });

    it("clicking 'Next' advances to step 2 (BillingPayment)", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      expect(screen.getByTestId("billing-payment")).toBeInTheDocument();
    });

    it("announces step 1 in the sr-only live region", () => {
      renderDialog();
      expect(screen.getByRole("status")).toHaveTextContent(
        /step 1 of 4.*overview/i,
      );
    });
  });

  describe("Step 2 — Payment method", () => {
    it("renders BillingPayment on step 2", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      expect(screen.getByTestId("billing-payment")).toBeInTheDocument();
    });

    it("'Next' button is disabled until onHasDefault fires", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      expect(screen.getByRole("button", { name: /^next$/i })).toBeDisabled();
    });

    it("'Next' button is enabled after onHasDefault fires", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      await user.click(screen.getByTestId("trigger-has-default"));
      expect(
        screen.getByRole("button", { name: /^next$/i }),
      ).not.toBeDisabled();
    });

    it("'Next' becomes disabled again when onNoPaymentMethods fires", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      await user.click(screen.getByTestId("trigger-has-default"));
      await user.click(screen.getByTestId("trigger-no-payment-methods"));
      expect(screen.getByRole("button", { name: /^next$/i })).toBeDisabled();
    });

    it("clicking 'Next' after onHasDefault advances to step 3 (BillingCheckout)", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      expect(screen.getByTestId("billing-checkout")).toBeInTheDocument();
    });

    it("announces step 2 in the sr-only live region", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      expect(screen.getByRole("status")).toHaveTextContent(
        /step 2 of 4.*payment method/i,
      );
    });
  });

  describe("Step 3 — Review", () => {
    it("renders BillingCheckout on step 3", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      expect(screen.getByTestId("billing-checkout")).toBeInTheDocument();
    });

    it("has a 'Confirm subscription' button on step 3", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      expect(
        screen.getByRole("button", { name: /confirm subscription/i }),
      ).toBeInTheDocument();
    });

    it("'Confirm subscription' calls createSubscription.mutateAsync", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() => expect(mockMutateAsync).toHaveBeenCalledWith({}));
    });

    it("advances to step 4 after subscription is active", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      mockRefetchSubscription.mockResolvedValue({ data: { status: "active" } });
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByTestId("billing-successful")).toBeInTheDocument(),
      );
    });

    it("advances to step 4 when subscription status is 'trialing'", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      mockRefetchSubscription.mockResolvedValue({
        data: { status: "trialing" },
      });
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByTestId("billing-successful")).toBeInTheDocument(),
      );
    });

    it("shows error and stays on step 3 when status is 'incomplete'", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      mockRefetchSubscription.mockResolvedValue({
        data: { status: "incomplete" },
      });
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /additional confirmation/i,
        ),
      );
      expect(screen.getByTestId("billing-checkout")).toBeInTheDocument();
    });

    it("shows 'wasn't fully activated' error for non-active non-incomplete statuses", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      mockRefetchSubscription.mockResolvedValue({
        data: { status: "past_due" },
      });
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /wasn't fully activated/i,
        ),
      );
    });

    it("shows 'unpaid invoices' error on 402 response", async () => {
      const err = { status: 402 };
      mockMutateAsync.mockRejectedValue(err);
      mockIsSdkError.mockImplementation((e: unknown) => e === err);
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(/unpaid invoices/i),
      );
    });

    it("shows generic error on non-402 failure", async () => {
      mockMutateAsync.mockRejectedValue(new Error("network failure"));
      mockIsSdkError.mockReturnValue(false);
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByRole("alert")).toHaveTextContent(
          /couldn't complete your subscription/i,
        ),
      );
    });

    it("disables 'Confirm subscription' and shows 'Subscribing…' while pending", async () => {
      mockIsPending.value = true;
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      const btn = screen.getByRole("button", { name: /subscribing/i });
      expect(btn).toBeDisabled();
    });

    it("announces step 3 in the sr-only live region", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      expect(screen.getByRole("status")).toHaveTextContent(
        /step 3 of 4.*review/i,
      );
    });
  });

  describe("Step 4 — Success", () => {
    async function goToStep4(user: ReturnType<typeof userEvent.setup>) {
      mockMutateAsync.mockResolvedValue(undefined);
      mockRefetchSubscription.mockResolvedValue({ data: { status: "active" } });
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByTestId("billing-successful")).toBeInTheDocument(),
      );
    }

    it("renders BillingSuccessful on step 4", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep4(user);
      expect(screen.getByTestId("billing-successful")).toBeInTheDocument();
    });

    it("clicking 'Done' calls onSuccess and onClose", async () => {
      const user = userEvent.setup();
      const { onClose, onSuccess } = renderDialog();
      await goToStep4(user);
      await user.click(screen.getByRole("button", { name: /done/i }));
      expect(onSuccess).toHaveBeenCalledOnce();
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("announces step 4 in the sr-only live region", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep4(user);
      expect(screen.getByRole("status")).toHaveTextContent(
        /step 4 of 4.*success/i,
      );
    });
  });

  describe("Navigation — Back button", () => {
    it("'Back' on step 2 returns to step 1", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      await user.click(screen.getByRole("button", { name: /back/i }));
      expect(screen.getByTestId("billing-letter")).toBeInTheDocument();
    });

    it("'Back' on step 3 returns to step 2", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(screen.getByRole("button", { name: /back/i }));
      expect(screen.getByTestId("billing-payment")).toBeInTheDocument();
    });

    it("'Back' clears any existing error message", async () => {
      mockMutateAsync.mockResolvedValue(undefined);
      mockRefetchSubscription.mockResolvedValue({
        data: { status: "past_due" },
      });
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      await user.click(
        screen.getByRole("button", { name: /confirm subscription/i }),
      );
      await waitFor(() =>
        expect(screen.getByRole("alert")).toBeInTheDocument(),
      );
      await user.click(screen.getByRole("button", { name: /back/i }));
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("Close / X button", () => {
    it("clicking 'Close' in the footer calls onClose on step 1", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await user.click(screen.getByRole("button", { name: /^close$/i }));
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("clicking the X (aria-label 'Close wizard') calls onClose on step 1", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog();
      await user.click(screen.getByRole("button", { name: /close wizard/i }));
      expect(onClose).toHaveBeenCalledOnce();
    });

    it("close buttons are present on step 2", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep2(user);
      expect(
        screen.getByRole("button", { name: /close wizard/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /^close$/i }),
      ).toBeInTheDocument();
    });

    it("close buttons are present on step 3", async () => {
      const user = userEvent.setup();
      renderDialog();
      await goToStep3(user);
      expect(
        screen.getByRole("button", { name: /close wizard/i }),
      ).toBeInTheDocument();
    });
  });
});
