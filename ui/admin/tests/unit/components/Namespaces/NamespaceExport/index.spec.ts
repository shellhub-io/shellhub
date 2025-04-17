import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper, DOMWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { INotificationsSuccess, INotificationsError } from "@admin/interfaces/INotifications";
import { saveAs } from "file-saver";
import NamespaceExport from "../../../../../src/components/Namespace/NamespaceExport.vue";

vi.mock("file-saver", () => ({
  saveAs: vi.fn(),
}));

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
    const snackbarStore = useSnackbarStore();

    vi.spyOn(namespaceStore, "setFilterNamespaces").mockResolvedValue(undefined);
    vi.spyOn(namespaceStore, "exportNamespacesToCsv").mockResolvedValue("csv_content");
    vi.spyOn(snackbarStore, "showSnackbarSuccessAction").mockImplementation(() => INotificationsSuccess.exportNamespaces);
    vi.spyOn(snackbarStore, "showSnackbarErrorAction").mockImplementation(() => INotificationsError.exportNamespaces);

    wrapper = mount(NamespaceExport, {
      attachTo: document.body,
      global: {
        plugins: [vuetify],
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

    expect(useSnackbarStore().showSnackbarSuccessAction).toHaveBeenCalledWith(INotificationsSuccess.exportNamespaces);
  });
});
