import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import { router } from "@/router";
import { SnackbarPlugin } from "@/plugins/snackbar";
import ConnectorEdit from "@/components/Connector/ConnectorEdit.vue";

type ConnectorEditWrapper = VueWrapper<InstanceType<typeof ConnectorEdit>>;

describe("Connector Edit", () => {
  let wrapper: ConnectorEditWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(ConnectorEdit, {
      global: {
        plugins: [vuetify, router, SnackbarPlugin],
      },
      props: {
        secure: true,
        uid: "fake-uid",
        ipAddress: "",
        portAddress: 80,
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
    expect(wrapper.find('[data-test="connector-edit-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="connector-edit-icon"]').exists()).toBe(true);
    await wrapper.findComponent('[data-test="connector-edit-btn"]').trigger("click");
    await flushPromises();
  });
});
