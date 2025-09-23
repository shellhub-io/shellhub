import MockAdapter from "axios-mock-adapter";
import { createPinia, setActivePinia } from "pinia";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach, afterEach, vi } from "vitest";
import DeviceRename from "@/components/Devices/DeviceRename.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import useDevicesStore from "@/store/modules/devices";
import { devicesApi } from "@/api/http";

describe("DeviceRename", () => {
  let wrapper: VueWrapper<InstanceType<typeof DeviceRename>>;
  setActivePinia(createPinia());
  const devicesStore = useDevicesStore();
  const vuetify = createVuetify();
  const mockDevicesApi = new MockAdapter(devicesApi.getAxios());

  beforeEach(async () => {
    wrapper = mount(DeviceRename, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        uid: "a582b47a42d",
        name: "39-5e-2a",
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders list item with rename option", () => {
    expect(wrapper.find('[data-test="rename-icon"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="rename-title"]').exists()).toBe(true);
  });

  it("Shows FormDialog with correct props when opened", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    expect(formDialog.exists()).toBe(true);
    expect(formDialog.props("title")).toBe("Rename Device");
    expect(formDialog.props("icon")).toBe("mdi-pencil");
    expect(formDialog.props("confirmText")).toBe("Rename");
    expect(formDialog.props("cancelText")).toBe("Close");
  });

  it("Shows text field with initial device name", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    const textField = wrapper.findComponent({ name: "VTextField" });
    expect(textField.exists()).toBe(true);
    expect(textField.props("modelValue")).toBe("39-5e-2a");
  });

  it("Calls renameDevice when rename is successful", async () => {
    const storeSpy = vi.spyOn(devicesStore, "renameDevice").mockResolvedValue();

    wrapper.vm.newName = "new-device-name";
    await wrapper.vm.rename();

    expect(storeSpy).toHaveBeenCalledWith({
      uid: "a582b47a42d",
      name: { name: "new-device-name" },
    });
  });

  it("Closes dialog when cancel is emitted", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();

    const formDialog = wrapper.findComponent({ name: "FormDialog" });
    await formDialog.vm.$emit("cancel");

    expect(wrapper.vm.showDialog).toBe(false);
  });

  it("Handles rename errors gracefully", async () => {
    mockDevicesApi.onPut("http://localhost:3000/api/devices/a582b47a42d").reply(400);

    const storeSpy = vi.spyOn(devicesStore, "renameDevice");
    wrapper.vm.newName = "new-name";
    await wrapper.vm.rename();

    expect(storeSpy).toHaveBeenCalled();
    // Component should handle errors gracefully without crashing
  });

  it("Exposes showDialog property", () => {
    expect(wrapper.vm.showDialog).toBeDefined();
    expect(typeof wrapper.vm.showDialog).toBe("boolean");
  });
});
