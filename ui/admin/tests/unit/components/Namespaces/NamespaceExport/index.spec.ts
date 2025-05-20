import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { saveAs } from "file-saver";
import { SnackbarInjectionKey } from "@/plugins/snackbar";
import NamespaceExport from "../../../../../src/components/Namespace/NamespaceExport.vue";

vi.mock("file-saver", () => ({
  saveAs: vi.fn(),
}));

const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
};

type NamespaceExportWrapper = VueWrapper<InstanceType<typeof NamespaceExport>>;

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

describe("NamespaceExport", () => {
  let wrapper: NamespaceExportWrapper;

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    setActivePinia(createPinia());
    const vuetify = createVuetify();

    const namespaceStore = useNamespacesStore();

    vi.spyOn(namespaceStore, "setFilterNamespaces").mockResolvedValue(undefined);
    vi.spyOn(namespaceStore, "exportNamespacesToCsv").mockResolvedValue("csv_content");

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
    expect(wrapper.find("[data-test='namespaces-export-btn']").exists()).toBe(true);
  });

  it("Opens the dialog and interacts with form", async () => {
    const dialog = new DOMWrapper(document.body);

    await wrapper.find("[data-test='namespaces-export-btn']").trigger("click");

    await flushPromises();

    await dialog.find("[data-test='form']").trigger("submit.prevent");

    await flushPromises();

    expect(useNamespacesStore().setFilterNamespaces).toHaveBeenCalled();
    expect(useNamespacesStore().exportNamespacesToCsv).toHaveBeenCalled();
    expect(saveAs).toHaveBeenCalled();

    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespaces exported successfully.");
  });
});
