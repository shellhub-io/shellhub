import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createVuetify } from "vuetify";
import ConnectorForm from "@/components/Connector/ConnectorForm.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ConnectorFormWrapper = VueWrapper<InstanceType<typeof ConnectorForm>>;

describe("Connector Form", () => {
  let wrapper: ConnectorFormWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(ConnectorForm, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        modelValue: true,
        isEditing: false,
        storeMethod: vi.fn(),
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("validates the address field", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const addressField = wrapper.findComponent('[data-test="address-text"]');

    await addressField.setValue("invalid ip");
    await flushPromises();
    expect(wrapper.findComponent('[data-test="address-text"]').text()).toContain("Invalid IP address format");
  });

  it("validates the port field", async () => {
    wrapper.vm.showDialog = true;
    await flushPromises();
    const addressField = wrapper.findComponent('[data-test="port-text"]');

    await addressField.setValue("invalid port");
    await flushPromises();
    expect(wrapper.findComponent('[data-test="port-text"]').text()).toContain("this must be a `number` type");
  });
});
