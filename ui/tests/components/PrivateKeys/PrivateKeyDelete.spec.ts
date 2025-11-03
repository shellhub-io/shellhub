import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PrivateKeyDelete from "@/components/PrivateKeys/PrivateKeyDelete.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";
import usePrivateKeysStore from "@/store/modules/private_keys";

type PrivateKeyDeleteWrapper = VueWrapper<InstanceType<typeof PrivateKeyDelete>>;

describe("Private Key Delete", () => {
  let wrapper: PrivateKeyDeleteWrapper;
  setActivePinia(createPinia());
  const privateKeysStore = usePrivateKeysStore();
  const vuetify = createVuetify();

  beforeEach(() => {
    wrapper = mount(PrivateKeyDelete, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: {
        id: 1,
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
    expect(wrapper.find('[data-test="privatekey-delete-btn"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="privatekey-delete-btn-title"]').exists()).toBe(true);

    await wrapper.find('[data-test="privatekey-delete-btn"]').trigger("click");
    await flushPromises();

    const body = new DOMWrapper(document.body);
    expect(body.find('[data-test="private-key-delete-dialog"]').exists()).toBe(true);

    const dialog = wrapper.findComponent({ name: "MessageDialog" });
    expect(dialog.exists()).toBe(true);
    expect(dialog.props("title")).toBe("Are you sure?");
    expect(dialog.props("description")).toBe("You are about to delete this private key");
    expect(dialog.props("icon")).toBe("mdi-alert");
    expect(dialog.props("iconColor")).toBe("error");
    expect(dialog.props("confirmText")).toBe("Delete");
    expect(dialog.props("confirmColor")).toBe("error");
    expect(dialog.props("cancelText")).toBe("Close");
  });

  it("Checks if the remove function updates the store on success", async () => {
    const storeSpy = vi.spyOn(privateKeysStore, "deletePrivateKey").mockResolvedValue();

    await wrapper.find('[data-test="privatekey-delete-btn"]').trigger("click");
    await flushPromises();

    const dialog = wrapper.findComponent({ name: "MessageDialog" });
    await dialog.vm.$emit("confirm");
    await flushPromises();

    expect(storeSpy).toHaveBeenCalledWith(1);
  });
});
