import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import SettingSessionRecording from "@/components/Setting/SettingSessionRecording.vue";
import { usersApi } from "@/api/http";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useSessionRecordingStore from "@/store/modules/session_recording";

type SettingSessionRecordingWrapper = VueWrapper<InstanceType<typeof SettingSessionRecording>>;

describe("Setting Session Recording", () => {
  let wrapper: SettingSessionRecordingWrapper;
  setActivePinia(createPinia());
  const sessionRecordingStore = useSessionRecordingStore();
  const mockUsersApi = new MockAdapter(usersApi.getAxios());
  const vuetify = createVuetify();

  beforeEach(() => {
    window.matchMedia = vi.fn().mockImplementation((query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    }));

    localStorage.setItem("tenant", "fake-tenant-data");

    mockUsersApi.onGet("http://localhost:3000/api/users/security").reply(200, true);
    sessionRecordingStore.isEnabled = true;

    wrapper = mount(SettingSessionRecording, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        tenantId: "fake-tenant-data",
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Changes status in store when ref is mutated", () => {
    mockUsersApi.onPut("http://localhost:3000/api/users/security/fake-tenant-data").reply(200);

    const storeSpy = vi.spyOn(sessionRecordingStore, "setStatus");
    wrapper.vm.isSessionRecordingEnabled = false;

    expect(storeSpy).toHaveBeenCalledWith({ id: "fake-tenant-data", status: false });
  });
});
