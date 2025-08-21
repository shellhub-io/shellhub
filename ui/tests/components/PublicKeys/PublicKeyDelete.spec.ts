import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
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

type PublicKeyDeleteWrapper = VueWrapper<InstanceType<typeof PublicKeyDelete>>;

describe("Public Key Delete", () => {
  let wrapper: PublicKeyDeleteWrapper;
  setActivePinia(createPinia());
  const publicKeysStore = usePublicKeysStore();
  const vuetify = createVuetify();
  const mockSshApi = new MockAdapter(sshApi.getAxios());

  beforeEach(async () => {
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

  it("Renders components", async () => {
    expect(wrapper.find('[data-test="public-key-remove-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="remove-title"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="public-key-remove-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    await flushPromises();
    expect(dialog.find('[data-test="text-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="remove-btn"]').exists()).toBe(true);
  });

  it("Successfully removes a Public Key", async () => {
    await wrapper.findComponent('[data-test="public-key-remove-btn"]').trigger("click");
    mockSshApi.onDelete("http://localhost:3000/api/sshkeys/public-keys/fake-fingerprint").reply(200);
    const storeSpy = vi.spyOn(publicKeysStore, "deletePublicKey");
    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    expect(storeSpy).toHaveBeenCalledWith("fake-fingerprint");
  });

  it("Shows error snackbar if removing a Public Key fails", async () => {
    await wrapper.findComponent('[data-test="public-key-remove-btn"]').trigger("click");
    mockSshApi.onDelete("http://localhost:3000/api/sshkeys/public-keys/fake-fingerprint").reply(404); // non-existent key
    await wrapper.findComponent('[data-test="remove-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to remove the public key.");
  });
});
