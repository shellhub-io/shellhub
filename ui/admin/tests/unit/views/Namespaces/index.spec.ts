import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { SnackbarPlugin } from "@/plugins/snackbar";
import routes from "../../../../src/router";
import Namespaces from "../../../../src/views/Namespaces.vue";

type NamespacesWrapper = VueWrapper<InstanceType<typeof Namespaces>>;

describe("Namespaces", () => {
  let wrapper: NamespacesWrapper;

  beforeEach(async () => {
    const pinia = createPinia();
    setActivePinia(pinia);

    const namespacesStore = useNamespacesStore();
    vi.spyOn(namespacesStore, "getPerPage", "get").mockReturnValue(10);
    vi.spyOn(namespacesStore, "getPage", "get").mockReturnValue(1);
    vi.spyOn(namespacesStore, "getnumberOfNamespaces", "get").mockReturnValue(1);

    namespacesStore.search = vi.fn();
    namespacesStore.fetch = vi.fn();

    const snackbarStore = useSnackbarStore();
    snackbarStore.showSnackbarErrorAction = vi.fn();

    const vuetify = createVuetify();

    wrapper = mount(Namespaces, {
      global: {
        plugins: [pinia, vuetify, routes, SnackbarPlugin],
      },
    });

    await flushPromises();
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
    expect(wrapper.find("h1").text()).toContain("Namespaces");
    expect(wrapper.find("[data-test='namespaces-list']").exists()).toBe(true);
    expect(wrapper.find("[data-test='namespaces-export-btn']").exists()).toBe(true);
    expect(wrapper.find("input").exists()).toBe(true);
  });
});
