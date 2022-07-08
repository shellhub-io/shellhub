import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SettingPrivateKeys from "../../../src/components/Setting/SettingPrivateKeys.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

describe("SettingPrivateKeys", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  const numberPrivateKeys = 2;

  const privateKeys = [
    {
      name: "shellhub",
      data: "BBGVvbmF",
    },
    {
      name: "shellhub",
      data: "AbGVvbmF",
    },
  ];

  const headers = [
    {
      text: "Name",
      value: "name",
      align: "center",
      sortable: true,
    },
    {
      text: "Fingerprint",
      value: "data",
      align: "center",
      sortable: true,
    },
    {
      text: "Actions",
      value: "actions",
      align: "center",
      sortable: false,
    },
  ];

  const store = createStore({
    state: {
      privateKeys,
      numberPrivateKeys,
    },
    getters: {
      "privatekeys/list": (state) => state.privateKeys,
      "privatekeys/getNumberPrivateKeys": (state) => state.numberPrivateKeys,
    },
    actions: {},
  });

  ///////
  // In this case, the first time the user enters the private keys
  // tab, a dialog appears with a certain message.
  ///////

  describe("Dialog is true", () => {
    beforeEach(() => {
      wrapper = mount(SettingPrivateKeys, {
        global: {
          plugins: [[store, key], routes, vuetify],
        },
        shallow: true,
      });
    });

    ///////
    // Component Rendering
    //////

    it("Is a Vue instance", () => {
      expect(wrapper).toBeTruthy();
    });
    it("Renders the component", () => {
      expect(wrapper.html()).toMatchSnapshot();
    });

    ///////s
    // Data checking
    //////
    it("Data is defined", () => {
      expect(wrapper.vm.$data).toBeDefined();
    });
  });
});
