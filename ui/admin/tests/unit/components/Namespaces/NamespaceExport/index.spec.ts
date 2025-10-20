import { createVuetify } from "vuetify";
import { flushPromises, mount, DOMWrapper } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";
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

describe("NamespaceExport", () => {
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();

  vi.spyOn(namespacesStore, "exportNamespacesToCsv").mockResolvedValue("csv_content");

  const wrapper = mount(NamespaceExport, {
    attachTo: document.body,
    global: {
      plugins: [createVuetify()],
      provide: { [SnackbarInjectionKey]: mockSnackbar },
    },
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
    const dialog = new DOMWrapper(document.body);
    expect(dialog.html()).toMatchSnapshot();
  });

  it("Opens the dialog and interacts with form", async () => {
    const dialog = new DOMWrapper(document.body);
    await wrapper.find("[data-test='namespaces-export-btn']").trigger("click");
    await flushPromises();
    await dialog.find("[data-test='confirm-btn']").trigger("click");
    await flushPromises();
    expect(useNamespacesStore().exportNamespacesToCsv).toHaveBeenCalled();
    expect(saveAs).toHaveBeenCalled();
    expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Namespaces exported successfully.");
  });
});
