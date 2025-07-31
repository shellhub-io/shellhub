import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import ConnectorAdd from "@/components/Connector/ConnectorAdd.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";

type ConnectorAddWrapper = VueWrapper<InstanceType<typeof ConnectorAdd>>;

describe("Connector Add", () => {
  let wrapper: ConnectorAddWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(ConnectorAdd, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
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
    await wrapper.findComponent('[data-test="connector-add-btn"]').trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(wrapper.find('[data-test="connector-add-btn"]').exists()).toBe(true);
    await flushPromises();
    expect(dialog.findComponent('[data-test="connector-form-component"]').exists());
  });
});
