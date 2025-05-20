import { beforeEach, describe, it, expect, vi, afterEach } from "vitest";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import useSnackbarStore from "@admin/store/modules/snackbar";
import { INotificationsCopy } from "@admin/interfaces/INotifications";
import { SnackbarPlugin } from "@/plugins/snackbar";
import UserResetPassword from "../../../../../src/components/User/UserResetPassword.vue";

type UserResetPasswordWrapper = VueWrapper<InstanceType<typeof UserResetPassword>>;

const node = document.createElement("div");
node.setAttribute("id", "app");
document.body.appendChild(node);

describe("User Reset Password", () => {
  let wrapper: UserResetPasswordWrapper;
  const vuetify = createVuetify();
  const mockProps = { userId: "user123" };

  beforeEach(() => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    setActivePinia(createPinia());

    const userStore = useUsersStore();
    const snackbarStore = useSnackbarStore();

    userStore.generatedPassword = "mocked-password";
    vi.spyOn(userStore, "resetUserPassword").mockResolvedValue(undefined);
    vi.spyOn(userStore, "refresh").mockResolvedValue(undefined);
    vi.spyOn(snackbarStore, "showSnackbarCopy").mockImplementation(() => INotificationsCopy.tenantId);
    vi.spyOn(snackbarStore, "showSnackbarErrorAction").mockImplementation(() => "Failed to reset user password. Please try again.");

    wrapper = mount(UserResetPassword, {
      global: {
        plugins: [vuetify, SnackbarPlugin],
      },
      props: mockProps,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
    wrapper.unmount();
  });

  it("is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("renders correctly", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("opens the dialog when the icon is clicked", async () => {
    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    expect(wrapper.vm.dialog).toBe(true);
  });

  it("closes the dialog and resets step", async () => {
    const dialog = new DOMWrapper(document.body);
    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    wrapper.vm.step = "step-2";
    await flushPromises();

    await dialog.find("[data-test='close-btn']").trigger("click");

    expect(wrapper.vm.dialog).toBe(false);
    expect(wrapper.vm.step).toBe("step-1");
  });

  it("proceeds to step 2 after clicking 'Enable'", async () => {
    const dialog = new DOMWrapper(document.body);
    const userStore = useUsersStore();

    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    await dialog.find("[data-test='enable-btn']").trigger("click");
    await flushPromises();

    expect(userStore.resetUserPassword).toHaveBeenCalledWith(mockProps.userId);
    expect(wrapper.vm.step).toBe("step-2");
  });

  it("shows an error when resetUserPassword fails", async () => {
    const dialog = new DOMWrapper(document.body);

    const userStore = useUsersStore();
    const snackbarStore = useSnackbarStore();

    vi.spyOn(userStore, "resetUserPassword").mockRejectedValueOnce(new Error("Failure"));

    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    await dialog.find("[data-test='enable-btn']").trigger("click");
    await flushPromises();

    expect(snackbarStore.showSnackbarErrorAction).toHaveBeenCalled();
    expect(wrapper.vm.step).toBe("step-1");
  });

  it("copies password to clipboard when clicked", async () => {
    const dialog = new DOMWrapper(document.body);

    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn(),
      },
    });

    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    await dialog.find("[data-test='enable-btn']").trigger("click");
    await flushPromises();

    await dialog.find("[data-test='generated-password-field']").trigger("click");

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith("mocked-password");
  });
});
