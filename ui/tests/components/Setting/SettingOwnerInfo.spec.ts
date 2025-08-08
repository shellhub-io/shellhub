import { setActivePinia, createPinia } from "pinia";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";
import SettingOwnerInfo from "@/components/Setting/SettingOwnerInfo.vue";
import { store, key } from "@/store";
import { SnackbarPlugin } from "@/plugins/snackbar";

type SettingOwnerInfoWrapper = VueWrapper<InstanceType<typeof SettingOwnerInfo>>;

describe("Setting Owner Info", () => {
  let wrapper: SettingOwnerInfoWrapper;
  setActivePinia(createPinia());
  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(SettingOwnerInfo, {
      global: {
        plugins: [[store, key], vuetify, SnackbarPlugin],
      },
      props: {
        isOwner: false,
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Displays message when user is not the owner", async () => {
    expect(wrapper.find('[data-test="message-div"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="contactUser-p"]').text()).toContain("Contact  user for more information.");
  });
});
