import { ref } from "vue";
import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { VLayout } from "vuetify/components";
import SettingDrawer from "@/components/Setting/SettingDrawer.vue";
import { store, key } from "@/store";
import { router } from "@/router";

const Component = {
  template: "<v-layout><SettingDrawer v-model='show'/></v-layout>",
};

describe("Setting Drawer", () => {
  let wrapper: VueWrapper<unknown>;

  const vuetify = createVuetify();

  beforeEach(async () => {
    wrapper = mount(Component, {
      global: {
        plugins: [[store, key], vuetify, router],
        config: {
          errorHandler: () => { /* ignore global error handler */ },
        },
        components: {
          "v-layout": VLayout,
          SettingDrawer,
        },
      },
      data() {
        return { show: ref(true) };
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("Is a Vue instance", () => {
    expect(wrapper.vm).toBeTruthy();
  });

  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("renders the v-list with correct items", () => {
    const listItems = wrapper.findAll(".v-list-item");
    const visibleItems = wrapper.findComponent(SettingDrawer)?.vm.visibleItems;

    expect(listItems.length).toBe(visibleItems.length);

    listItems.forEach((item, index) => {
      expect(item.text()).toBe(visibleItems[index].title);
    });
  });

  it("toggles the drawer when showNavigationDrawer changes", async () => {
    const drawer = wrapper.findComponent(SettingDrawer);

    expect(drawer.find(".v-navigation-drawer--active").exists()).toBe(true);

    await wrapper.setData({ show: false });

    expect(drawer.find(".v-navigation-drawer--active").exists()).toBe(false);
  });
});
