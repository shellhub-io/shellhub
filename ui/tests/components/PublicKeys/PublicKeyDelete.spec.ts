import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import PublicKeyDelete from "@/components/PublicKeys/PublicKeyDelete.vue";
import { sshApi } from "@/api/http";
import { router } from "@/router";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import usePublicKeysStore from "@/store/modules/public_keys";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Public Key Delete", () => {
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const mockSshApi = new MockAdapter(sshApi.getAxios());
  const publicKeysStore = usePublicKeysStore();

  let wrapper: ReturnType<typeof mount<InstanceType<typeof PublicKeyDelete>>>;

  beforeEach(() => {
    vi.clearAllMocks();
    wrapper = mount(PublicKeyDelete, {
      global: {
        plugins: [vuetify, router],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        fingerprint: "fake-fingerprint",
        hasAuthorization: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders components (MessageDialog props)", async () => {
    await wrapper.find('[data-test="public-key-remove-btn"]').trigger("click");
    await flushPromises();

    const dlg = wrapper.findComponent({ name: "MessageDialog" });
    expect(dlg.exists()).toBe(true);
    expect(dlg.props("title")).toBe("Are you sure?");
    expect(dlg.props("description")).toBe("You are about to delete this public key");
    expect(dlg.props("icon")).toBe("mdi-alert");
    expect(dlg.props("iconColor")).toBe("error");
    expect(dlg.props("confirmText")).toBe("Delete");
    expect(dlg.props("confirmColor")).toBe("error");
    expect(dlg.props("cancelText")).toBe("Close");
  });

  it("Successfully removes a Public Key", async () => {
    mockSshApi
      .onDelete("http://localhost:3000/api/sshkeys/public-keys/fake-fingerprint")
      .reply(200);

    const storeSpy = vi.spyOn(publicKeysStore, "deletePublicKey").mockResolvedValue();

    await wrapper.find('[data-test="public-key-remove-btn"]').trigger("click");
    const dlg = wrapper.findComponent({ name: "MessageDialog" });
    await dlg.vm.$emit("confirm");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith("fake-fingerprint");
  });

  it("Shows error snackbar if removing a Public Key fails", async () => {
    mockSshApi
      .onDelete("http://localhost:3000/api/sshkeys/public-keys/fake-fingerprint")
      .reply(404);

    vi.spyOn(publicKeysStore, "deletePublicKey").mockRejectedValue(new Error("not found"));

    await wrapper.find('[data-test="public-key-remove-btn"]').trigger("click");
    const dlg = wrapper.findComponent({ name: "MessageDialog" });
    await dlg.vm.$emit("confirm");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove the public key.");
  });
});
