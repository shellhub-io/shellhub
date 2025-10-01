import { createPinia, setActivePinia } from "pinia";
import { DOMWrapper, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { expect, describe, it } from "vitest";
import ContainerAdd from "@/components/Containers/ContainerAdd.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("ContainerAdd", () => {
  setActivePinia(createPinia());
  const vuetify = createVuetify();
  const wrapper = mount(ContainerAdd, {
    global: { plugins: [vuetify, SnackbarPlugin] },
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", async () => {
    const button = wrapper.find('[data-test="container-add-btn"]');
    await button.trigger("click");
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });
});
