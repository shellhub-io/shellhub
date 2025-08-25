import { createPinia, setActivePinia } from "pinia";
import { flushPromises, DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import MockAdapter from "axios-mock-adapter";
import { expect, describe, it, beforeEach, vi } from "vitest";
import TagFormUpdate from "@/components/Tags/TagFormUpdate.vue";
import { devicesApi, tagsApi } from "@/api/http";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import useDevicesStore from "@/store/modules/devices";

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

const devices = [
  {
    uid: "a582b47a42d",
    name: "39-5e-2a",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T18:58:53.276Z",
    online: false,
    namespace: "user",
    status: "accepted",
    tags: ["test"],
  },
  {
    uid: "a582b47a42e",
    name: "39-5e-2b",
    identity: {
      mac: "00:00:00:00:00:00",
    },
    info: {
      id: "linuxmint",
      pretty_name: "Linux Mint 19.3",
      version: "",
    },
    public_key: "----- PUBLIC KEY -----",
    tenant_id: "fake-tenant-data",
    last_seen: "2020-05-20T19:58:53.276Z",
    online: true,
    namespace: "user",
    status: "accepted",
    tags: ["test2"],
  },
];

describe("Tag Form Update", async () => {
  let wrapper: VueWrapper<InstanceType<typeof TagFormUpdate>>;
  setActivePinia(createPinia());
  const devicesStore = useDevicesStore();
  const vuetify = createVuetify();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());
  const mockTagsApi = new MockAdapter(tagsApi.getAxios());

  beforeEach(async () => {
    mockDevicesApi.onGet("http://localhost:3000/api/devices?page=1&per_page=10&status=accepted").reply(200, devices);
    mockTagsApi.onGet("http://localhost:3000/api/tags").reply(200, ["test2"]);

    wrapper = mount(TagFormUpdate, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
      props: {
        deviceUid: devices[0].uid,
        tagsList: devices[0].tags,
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

  it("Renders the component data table", async () => {
    const dialog = new DOMWrapper(document.body);
    await wrapper.setProps({ deviceUid: devices[0].uid, tagsList: devices[0].tags });
    await flushPromises();
    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    expect(wrapper.find('[data-test="has-tags-verification"]').exists()).toBe(true);
    expect(dialog.find('[data-test="title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="deviceTag-combobox"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="save-btn"]').exists()).toBe(true);
  });

  it("Successfully add tags", async () => {
    await wrapper.setProps({ deviceUid: devices[0].uid, tagsList: devices[0].tags });
    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d/tags").reply(200);
    const dialog = new DOMWrapper(document.body);
    const storeSpy = vi.spyOn(devicesStore, "updateDeviceTags");

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="deviceTag-combobox"').setValue(["tag-test"]);
    await dialog.find('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(storeSpy).toHaveBeenCalledWith({
      uid: "a582b47a42d",
      tags: { tags: ["test"] },
    });
  });

  it("Failed to add tags", async () => {
    await wrapper.setProps({ deviceUid: devices[0].uid, tagsList: devices[0].tags });
    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d/tags").reply(403);
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="open-tags-btn"]').trigger("click");
    await wrapper.findComponent('[data-test="deviceTag-combobox"').setValue(["tag-test"]);
    await dialog.find('[data-test="save-btn"]').trigger("click");
    await flushPromises();
    expect(mockSnackbar.showError).toHaveBeenCalledWith("You are not authorized to update this tag.");
  });
});
