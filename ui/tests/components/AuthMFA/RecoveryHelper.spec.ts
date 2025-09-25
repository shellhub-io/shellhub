import { createPinia, setActivePinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, afterEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { mfaApi } from "@/api/http";
import RecoveryHelper from "@/components/AuthMFA/RecoveryHelper.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useAuthStore from "@/store/modules/auth";

type RecoveryHelperWrapper = VueWrapper<InstanceType<typeof RecoveryHelper>>;

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

describe("Recovery Helper", () => {
  let wrapper: RecoveryHelperWrapper;
  const vuetify = createVuetify();
  setActivePinia(createPinia());
  const authStore = useAuthStore();
  const mockMfaApi = new MockAdapter(mfaApi.getAxios());

  beforeEach(() => {
    wrapper = mount(RecoveryHelper, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: { modelValue: true },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Disable MFA Authentication", async () => {
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(200);
    const mfaSpy = vi.spyOn(authStore, "disableMfa");

    await wrapper.findComponent('[data-test="disable-btn"]').trigger("click");

    expect(mfaSpy).toHaveBeenCalledWith({ recovery_code: "" });
    expect(authStore.isMfaEnabled).toBe(false);
  });

  it("Disable MFA Authentication (fail)", async () => {
    mockMfaApi.onPut("http://localhost:3000/api/user/mfa/disable").reply(403);

    await wrapper.findComponent('[data-test="disable-btn"]').trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("An error occurred while disabling MFA.");
  });
});
