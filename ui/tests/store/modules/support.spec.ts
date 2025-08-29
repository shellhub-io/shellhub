import { createPinia, setActivePinia } from "pinia";
import { describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { namespacesApi } from "@/api/http";
import useSupportStore from "@/store/modules/support";

vi.mock("@productdevbook/chatwoot/vue", () => ({
  useChatWoot: () => ({
    reset: vi.fn(),
  }),
}));

describe("Support Store", () => {
  const mockNamespacesApi = new MockAdapter(namespacesApi.getAxios());
  setActivePinia(createPinia());
  const supportStore = useSupportStore();

  it("should have initial state values", () => {
    expect(supportStore.identifier).toEqual("");
    expect(supportStore.isChatCreated).toEqual(false);
  });

  it("successfully gets identifier", async () => {
    const mockResponse = { identifier: "fake-identifier" };
    mockNamespacesApi.onGet("http://localhost:3000/api/namespaces/fake-tenant/support").reply(200, mockResponse);

    await supportStore.getIdentifier("fake-tenant");

    expect(supportStore.identifier).toEqual("fake-identifier");
  });
});
