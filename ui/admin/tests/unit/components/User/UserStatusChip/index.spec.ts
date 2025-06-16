import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import UserStatusChip from "@admin/components/User/UserStatusChip.vue";
import { UserStatus } from "@admin/interfaces/IUser";
import { SnackbarPlugin } from "@/plugins/snackbar";

describe("User Status Chip", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserStatusChip>>;
  const vuetify = createVuetify();

  const createWrapper = (status: UserStatus) => mount(UserStatusChip, {
    global: {
      plugins: [vuetify, SnackbarPlugin],
    },
    props: { status },
  });

  beforeEach(async () => {
    wrapper = createWrapper("confirmed");
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("renders confirmed status correctly", () => {
    wrapper = createWrapper("confirmed");
    const chip = wrapper.findComponent({ name: "VChip" });

    expect(chip.props("color")).toBe("success");
    expect(chip.props("prependIcon")).toBe("mdi-checkbox-marked-circle");
    expect(chip.text()).toBe("Confirmed");
  });

  it("renders invited status correctly", () => {
    wrapper = createWrapper("invited");
    const chip = wrapper.findComponent({ name: "VChip" });

    expect(chip.props("color")).toBe("warning");
    expect(chip.props("prependIcon")).toBe("mdi-email-alert");
    expect(chip.text()).toBe("Invited");
  });

  it("renders not-confirmed status correctly", () => {
    wrapper = createWrapper("not-confirmed");
    const chip = wrapper.findComponent({ name: "VChip" });

    expect(chip.props("color")).toBe("error");
    expect(chip.props("prependIcon")).toBe("mdi-alert-circle");
    expect(chip.text()).toBe("Not Confirmed");
  });

  it("handles invalid status by falling back to not-confirmed", () => {
    const wrapper = createWrapper("fake-status" as UserStatus);
    const chip = wrapper.findComponent({ name: "VChip" });

    expect(chip.props("color")).toBe("error");
    expect(chip.props("prependIcon")).toBe("mdi-alert-circle");
    expect(chip.text()).toBe("Not Confirmed");
  });
});
