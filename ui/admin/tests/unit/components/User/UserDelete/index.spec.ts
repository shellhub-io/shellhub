import { nextTick } from "vue";
import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserDelete from "@admin/components/User/UserDelete.vue";
import routes from "@admin/router";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("User Delete", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserDelete>>;
  setActivePinia(createPinia());
  const usersStore = useUsersStore();
  usersStore.deleteUser = vi.fn();
  usersStore.fetchUsersList = vi.fn();

  const createWrapper = (props = {}, slots = {}) => {
    return mount(UserDelete, {
      props: {
        id: "test-id",
        ...props,
      },
      slots: {
        default: "<button>Delete</button>",
        ...slots,
      },
      global: { plugins: [createVuetify(), routes, SnackbarPlugin] },
    });
  };

  afterEach(() => { wrapper.unmount(); });

  it("Exposes openDialog via slot props", () => {
    const slotMock = vi.fn();
    wrapper = createWrapper({}, { default: slotMock });

    expect(slotMock).toHaveBeenCalled();
    const slotProps = slotMock.mock.calls[0][0];
    expect(slotProps.openDialog).toBeInstanceOf(Function);
  });

  it("Shows tooltip text when enabled", () => {
    wrapper = createWrapper({ showTooltip: true });

    const tooltip = wrapper.findComponent({ name: "VTooltip" });
    expect(tooltip.props("text")).toBe("Remove");
  });

  it("Opens dialog when openDialog is called", async () => {
    const slotMock = vi.fn();
    wrapper = createWrapper({}, { default: slotMock });

    const slotProps = slotMock.mock.calls[0][0];

    slotProps.openDialog();
    await nextTick();

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    expect(messageDialog.props("modelValue")).toBe(true);
  });

  it("Deletes user on confirm", async () => {
    wrapper = createWrapper();

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    await messageDialog.vm.$emit("confirm");
    await flushPromises();

    expect(usersStore.deleteUser).toHaveBeenCalledWith("test-id");
    expect(usersStore.fetchUsersList).toHaveBeenCalled();
  });

  it("Redirects after delete when redirect prop is true", async () => {
    const pushSpy = vi.spyOn(routes, "push");

    wrapper = createWrapper({ redirect: true });

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    await messageDialog.vm.$emit("confirm");
    await flushPromises();

    expect(pushSpy).toHaveBeenCalledWith("/users");
  });

  it("Emits update event after successful deletion", async () => {
    wrapper = createWrapper();

    const messageDialog = wrapper.findComponent({ name: "MessageDialog" });
    await messageDialog.vm.$emit("confirm");
    await flushPromises();

    expect(wrapper.emitted("update")).toBeTruthy();
  });
});
