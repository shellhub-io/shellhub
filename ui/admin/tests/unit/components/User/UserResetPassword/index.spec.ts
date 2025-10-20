import { describe, it, expect, vi } from "vitest";
import { DOMWrapper, flushPromises, mount } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import { createPinia, setActivePinia } from "pinia";
import useUsersStore from "@admin/store/modules/users";
import UserResetPassword from "@admin/components/User/UserResetPassword.vue";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

const mockSnackbar = {
  showInfo: vi.fn(),
  showError: vi.fn(),
};

describe("User Reset Password", () => {
  const mockProps = { userId: "user123" };
  setActivePinia(createPinia());
  const usersStore = useUsersStore();

  vi.spyOn(usersStore, "resetUserPassword").mockResolvedValue("mocked-password");
  const wrapper = mount(UserResetPassword, {
    global: {
      plugins: [createVuetify()],
      provide: { [SnackbarInjectionKey]: mockSnackbar },
    },
    props: mockProps,
  });

  const dialog = new DOMWrapper(document.body);

  it("renders correctly", async () => {
    expect(wrapper.html()).toMatchSnapshot();
    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    expect(dialog.html()).toMatchSnapshot();
  });

  it("closes the dialog and resets step", async () => {
    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    wrapper.vm.step = 2;
    await flushPromises();

    await dialog.find("[data-test='close-btn']").trigger("click");

    expect(wrapper.vm.showDialog).toBe(false);
    expect(wrapper.vm.step).toBe(1);
  });

  it("proceeds to step 2 after clicking 'Enable'", async () => {
    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    await dialog.find("[data-test='enable-btn']").trigger("click");
    await flushPromises();

    expect(usersStore.resetUserPassword).toHaveBeenCalledWith(mockProps.userId);
    expect(wrapper.vm.step).toBe(2);
  });

  it("shows an error when resetUserPassword fails", async () => {
    wrapper.vm.step = 1;

    vi.spyOn(usersStore, "resetUserPassword").mockRejectedValueOnce(new Error("Failure"));

    await wrapper.find("[data-test='open-dialog-icon']").trigger("click");
    await dialog.find("[data-test='enable-btn']").trigger("click");
    await flushPromises();

    expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to reset user password. Please try again.");
    expect(wrapper.vm.step).toBe(1);
  });
});
