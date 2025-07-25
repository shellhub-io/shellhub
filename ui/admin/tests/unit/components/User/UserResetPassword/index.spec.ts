import { beforeEach, describe, it, expect, vi, afterEach } from "vitest";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserResetPassword from "@admin/components/User/UserResetPassword.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

type UserResetPasswordWrapper = VueWrapper<InstanceType<typeof UserResetPassword>>;

const mockSnackbar = {
  showInfo: vi.fn(),
  showError: vi.fn(),
};

describe("User Reset Password", () => {
  let wrapper: UserResetPasswordWrapper;
  const vuetify = createVuetify();
  const mockProps = { userId: "user123" };

  beforeEach(() => {
    setActivePinia(createPinia());

    const userStore = useUsersStore();

    userStore.generatedPassword = "mocked-password";
    vi.spyOn(userStore, "resetUserPassword").mockResolvedValue(undefined);
    vi.spyOn(userStore, "refresh").mockResolvedValue(undefined);

    wrapper = mount(UserResetPassword, {
      global: {
        plugins: [vuetify],
        provide: { [SnackbarInjectionKey]: mockSnackbar },
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
    expect(wrapper.vm.showDialog).toBe(true);
  });

  it("closes the dialog and resets step", async () => {
    const dialog = new DOMWrapper(document.body);
    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    wrapper.vm.step = "step-2";
    await flushPromises();

    await dialog.find("[data-test='close-btn']").trigger("click");

    expect(wrapper.vm.showDialog).toBe(false);
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

    vi.spyOn(userStore, "resetUserPassword").mockRejectedValueOnce(new Error("Failure"));

    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    await dialog.find("[data-test='enable-btn']").trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to reset user password. Please try again.");
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
