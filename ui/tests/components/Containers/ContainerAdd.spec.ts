import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it, beforeEach } from "vitest";
import ContainerAdd from "@/components/Containers/ContainerAdd.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("ContainerAdd", () => {
  let wrapper: VueWrapper<InstanceType<typeof ContainerAdd>>;
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  beforeEach(async () => {
    wrapper = mount(ContainerAdd, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
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
    const button = wrapper.find('[data-test="container-add-btn"]');
    expect(button.exists()).toBe(true);
    await button.trigger("click");
    const dialog = new DOMWrapper(document.body);

    expect(dialog.find('[data-test="container-add-dialog"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="dialog-text"]').exists()).toBe(true);
    expect(dialog.find('[data-test="command-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="documentation-link"]').exists()).toBe(true);
    expect(dialog.find('[data-test="close-btn"]').exists()).toBe(true);
  });
});
