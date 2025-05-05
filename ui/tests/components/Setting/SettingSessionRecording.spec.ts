import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import SettingSessionRecording from "@/components/Setting/SettingSessionRecording.vue";
import { usersApi } from "@/api/http";
import { store, key } from "@/store";

type SettingSessionRecordingWrapper = VueWrapper<InstanceType<typeof SettingSessionRecording>>;

describe("Setting Session Recording", () => {
  let wrapper: SettingSessionRecordingWrapper;

  const vuetify = createVuetify();

  let mockUser: MockAdapter;

  beforeEach(async () => {
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
    mockUser = new MockAdapter(usersApi.getAxios());

    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, true);
    store.commit("sessionRecording/setEnabled", true);

    wrapper = mount(SettingSessionRecording, {
      global: {
        plugins: [[store, key], vuetify],
      },
      props: {
        hasTenant: true,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Changes status in store when ref is mutated", async () => {
    mockUser.onPut("http://localhost:3000/api/users/security/fake-tenant-data").reply(200);

    const dispatchSpy = vi.spyOn(store, "dispatch");
    wrapper.vm.sessionRecordingStatus = false;

    expect(dispatchSpy).toHaveBeenCalledWith("sessionRecording/setStatus", { id: "fake-tenant-data", status: false });
  });
});
