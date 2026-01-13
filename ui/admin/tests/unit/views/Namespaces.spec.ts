import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import routes from "@admin/router";
import Namespaces from "@admin/views/Namespaces.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

type NamespacesWrapper = VueWrapper<InstanceType<typeof Namespaces>>;

describe("Namespaces", () => {
  let wrapper: NamespacesWrapper;
  const pinia = createPinia();
  setActivePinia(pinia);
  const namespacesStore = useNamespacesStore();
  namespacesStore.fetchNamespaceList = vi.fn();
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(Namespaces, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Renders the template with default data", () => {
    expect(wrapper.vm.filter).toBe("");
  });

  it("Must change the filter value when input change", async () => {
    expect(wrapper.vm.filter).toBe("");
    const input = wrapper.find("input");
    await input.setValue("ShellHub");
    expect(wrapper.vm.filter).toBe("ShellHub");
  });

  it("Should render all the components in the screen", () => {
    expect(wrapper.find("[data-test='namespaces-list']").exists()).toBe(true);
    expect(wrapper.find("[data-test='namespaces-export-btn']").exists()).toBe(true);
    expect(wrapper.find("input").exists()).toBe(true);
  });
});
