import { createPinia, setActivePinia } from "pinia";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import useBillingStore from "@/store/modules/billing";
import { billingApi } from "@/api/http";
import axios from "axios";
import { envVariables } from "@/envVariables";
import { IBilling } from "@/interfaces/IBilling";

const mockBillingBase: IBilling = {
  id: "sub_123",
  active: true,
  status: "active",
  end_at: 1735689600,
  invoices: [
    { id: "inv_1", amount: 2999, status: "paid", currency: "usd" },
    { id: "inv_2", amount: 2999, status: "open", currency: "usd" },
  ],
};

describe("Billing Store", () => {
  let billingMock: MockAdapter;
  let axiosMock: MockAdapter;
  let store: ReturnType<typeof useBillingStore>;

  beforeEach(() => {
    setActivePinia(createPinia());
    billingMock = new MockAdapter(billingApi.getAxios());
    axiosMock = new MockAdapter(axios);
    store = useBillingStore();
    vi.stubGlobal("window", { open: vi.fn() });
    vi.spyOn(envVariables, "isCloud", "get").mockReturnValue(true);
  });

  afterEach(() => {
    billingMock.reset();
    axiosMock.reset();
    vi.unstubAllGlobals();
  });

  describe("Initial State", () => {
    it("should have correct default values", () => {
      expect(store.billing).toEqual({});
      expect(store.isActive).toBe(false);
      expect(store.status).toBe("inactive");
      expect(store.invoices).toEqual([]);
      expect(store.showBillingWarning).toBe(false);
    });
  });

  describe("showBillingWarning", () => {
    it("should allow setting showBillingWarning to true", () => {
      store.showBillingWarning = true;
      expect(store.showBillingWarning).toBe(true);
    });

    it("should allow setting showBillingWarning to false", () => {
      store.showBillingWarning = true;
      store.showBillingWarning = false;
      expect(store.showBillingWarning).toBe(false);
    });
  });

  describe("Computed Properties", () => {
    it("should compute isActive correctly when active is true", () => {
      store.billing = { ...mockBillingBase, active: true };

      expect(store.isActive).toBe(true);
    });

    it("should compute isActive correctly when active is false", () => {
      store.billing = { ...mockBillingBase, active: false };

      expect(store.isActive).toBe(false);
    });

    it("should compute isActive as false when active is undefined", () => {
      const { active: _active, ...billingWithoutActive } = mockBillingBase;
      store.billing = billingWithoutActive as IBilling;

      expect(store.isActive).toBe(false);
    });

    it("should compute status correctly", () => {
      store.billing = { ...mockBillingBase, status: "trialing" };

      expect(store.status).toBe("trialing");
    });

    it("should compute status as inactive when undefined", () => {
      const { status: _status, ...billingWithoutStatus } = mockBillingBase;
      store.billing = billingWithoutStatus as IBilling;

      expect(store.status).toBe("inactive");
    });

    it("should compute invoices correctly", () => {
      store.billing = mockBillingBase;

      expect(store.invoices).toEqual(mockBillingBase.invoices);
    });

    it("should compute invoices as empty array when undefined", () => {
      const { invoices: _invoices, ...billingWithoutInvoices } = mockBillingBase;
      store.billing = billingWithoutInvoices as IBilling;

      expect(store.invoices).toEqual([]);
    });
  });

  describe("getSubscriptionInfo", () => {
    const subscriptionUrl = "http://localhost:3000/api/billing/subscription";

    it("should fetch subscription info successfully when isCloud is true", async () => {
      billingMock.onGet(subscriptionUrl).reply(200, mockBillingBase);

      await store.getSubscriptionInfo();

      expect(store.billing).toEqual(mockBillingBase);
      expect(store.isActive).toBe(true);
      expect(store.status).toBe("active");
      expect(store.invoices).toEqual(mockBillingBase.invoices);
    });

    it("should not fetch subscription info when isCloud is false", async () => {
      vi.spyOn(envVariables, "isCloud", "get").mockReturnValue(false);

      // Should not make any API call
      billingMock.onGet(subscriptionUrl).reply(200, { active: true });

      await store.getSubscriptionInfo();

      expect(store.billing).toEqual({});
      expect(store.isActive).toBe(false);
    });

    it("should set active to false when request fails", async () => {
      billingMock.onGet(subscriptionUrl).reply(404);

      await store.getSubscriptionInfo();

      expect(store.billing.active).toBe(false);
    });

    it("should set active to false when network error occurs", async () => {
      billingMock.onGet(subscriptionUrl).networkError();

      await store.getSubscriptionInfo();

      expect(store.billing.active).toBe(false);
    });
  });

  describe("openBillingPortal", () => {
    const portalUrl = "/api/billing/portal";

    it("should open billing portal successfully", async () => {
      const mockUrl = "https://billing.stripe.com/session/xyz";

      axiosMock.onPost(portalUrl).reply(200, { url: mockUrl });

      await store.openBillingPortal();

      expect(window.open).toHaveBeenCalledWith(mockUrl, "_blank");
    });

    it("should throw error when request fails", async () => {
      axiosMock.onPost(portalUrl).reply(403);

      await expect(store.openBillingPortal()).rejects.toBeAxiosErrorWithStatus(403);
    });

    it("should throw error when network error occurs", async () => {
      axiosMock.onPost(portalUrl).networkError();

      await expect(store.openBillingPortal()).rejects.toThrow();
    });
  });
});
