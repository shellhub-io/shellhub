import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { saveAs } from "file-saver";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceExport from "@admin/components/Namespace/NamespaceExport.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

vi.mock("file-saver", () => ({
  saveAs: vi.fn(),
}));

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type NamespaceExportWrapper = VueWrapper<InstanceType<typeof NamespaceExport>>;

describe("NamespaceExport", () => {
  let wrapper: NamespaceExportWrapper;
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    vi.spyOn(namespacesStore, "exportNamespacesToCsv").mockResolvedValue("csv_content");

    wrapper = mount(NamespaceExport, {
      attachTo: document.body,
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Opens the dialog and interacts with form", async () => {
    const dialog = new DOMWrapper(document.body);

    await wrapper.find("[data-test='namespaces-export-btn']").trigger("click");

    await flushPromises();

    await dialog.find("[data-test='form']").trigger("submit.prevent");

    await flushPromises();

    expect(useNamespacesStore().exportNamespacesToCsv).toHaveBeenCalled();
    expect(saveAs).toHaveBeenCalled();

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespaces exported successfully.");
  });
});
