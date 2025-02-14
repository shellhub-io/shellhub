import { beforeEach, describe, it, expect, vi } from "vitest";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { createVuetify } from "vuetify";
import UserResetPassword from "../../../../../src/components/User/UserResetPassword.vue";
import { store, key } from "../../../../../src/store";

type UserResetPasswordWrapper = VueWrapper<InstanceType<typeof UserResetPassword>>;

describe("User Reset Password", () => {
  let wrapper: UserResetPasswordWrapper;

  const vuetify = createVuetify();

  const mockProps = {
    userId: "user123",
  };

  beforeEach(() => {
    wrapper = mount(UserResetPassword, {
      global: {
        plugins: [[store, key], vuetify],
      },
      props: mockProps,
    });
  });

  it("is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });

  it("renders correctly", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("opens the dialog when the icon is clicked", async () => {
    await wrapper.findComponent("[data-test='open-dialog-icon']").trigger("click");
    expect(wrapper.vm.dialog).toBe(true);
  });

  it("closes the dialog when 'Close' is clicked", async () => {
    await wrapper.findComponent("[data-test='open-dialog-icon']").trigger("click");
    await flushPromises();
    await wrapper.findComponent("[data-test='cancel-btn']").trigger("click");
    expect(wrapper.vm.dialog).toBe(false);
    expect(wrapper.vm.step).toBe("step-1");
  });

  it("proceeds to the second step when 'Enable' is clicked", async () => {
    await wrapper.findComponent("[data-test='open-dialog-icon']").trigger("click");

    const resetUserPasswordSpy = vi.spyOn(store, "dispatch").mockResolvedValueOnce({});

    await wrapper.findComponent("[data-test='enable-btn']").trigger("click");
    expect(resetUserPasswordSpy).toHaveBeenCalledWith("users/resetUserPassword", mockProps.userId);
    expect(wrapper.vm.step).toBe("step-2");
  });

  it("shows an error message if 'Enable' action fails", async () => {
    await wrapper.findComponent("[data-test='open-dialog-icon']").trigger("click");

    const resetUserPasswordSpy = vi.spyOn(store, "dispatch").mockRejectedValueOnce(new Error("Error"));

    await wrapper.findComponent("[data-test='enable-btn']").trigger("click");
    expect(resetUserPasswordSpy).toHaveBeenCalledWith("users/resetUserPassword", mockProps.userId);
    expect(wrapper.vm.step).toBe("step-1");
  });

  it("resets dialog state on 'Close'", async () => {
    await wrapper.findComponent("[data-test='open-dialog-icon']").trigger("click");

    wrapper.vm.step = "step-2";

    await wrapper.findComponent("[data-test='cancel-btn']").trigger("click");
    expect(wrapper.vm.dialog).toBe(false);
    expect(wrapper.vm.step).toBe("step-1");
  });
});
