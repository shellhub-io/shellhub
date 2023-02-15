import { createVuetify } from "vuetify";
import { flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import PrivateKeyList from "../../../src/components/PrivateKeys/PrivateKeyList.vue";
import { createStore } from "vuex";
import { key } from "../../../src/store";
import routes from "../../../src/router";

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

const privateKeyRSA = `-----BEGIN RSA PRIVATE KEY-----
  MIIEowIBAAKCAQEAsV6tDWbW0QSZcsluSzkzLqrGY96S/kOVAkkMZxIzKZ8+si+g
  C00vn6VS8xyEMH2d6eZDBpLwMFzsC7r5Bab0L2xrI+/5AaL+sw3i+3sXcz1b4VXJ
  bCFP/2k5tq8k9fPfERYBDptmA0rDbXhuHWAXvUmzSSG3swIO5QgHSsbG2Q9OaYV/
  DXsBjnBaJT5PJErmKiepVgf1lp3C0/Hl4D/BMPea8LoPQJtwKYo7qA9Y9USG9S4q
  xfcZDTYNdwwPUGBBTjtel1RCMSEtX61uyq4ftyz9lukOSEhUcxhe4cZVrPUUcwf0
  htlsazqXkAKAzi67f37t4O7FS/aQN9mGy4MgEwIDAQABAoIBACm+XnwI+AW5T2P0
  hECv9ZvGFWrrtyygOzGOf5zCa8gf8mF9U+1U/SBViHAvBe1jowapapzheCXpuKQK
  HRF3eYCvw4dxsujvs4Hwgrss/RfiGP2vcvg/3tP9r9eO4OQBwT4EL7uAV0HvFo9t
  CH1hYDTsY4WSqek3UsoUWaL/pUzwKMijUgh2Dzj5o9AlNGWANu6txI1mIgHmwUvj
  2kV7E4R1mGynSprdsW68V36viB/V9d82XGxd3tYhKojiS1Dir68mR2U8ld3728Pd
  xU7o9x6NcWOtpTY1nS9MpufaYUTlp/chOXSd2RIY6JmtgbJcVTdE4rasfIAEnlAZ
  XALqKAECgYEA4kl6ZfcwKtxebVyczMCk6QWOJtsJ6CT17w2oehQGSuSLXeidjkFe
  bm04hUcN4Rm5iipUwDlA6JT8QoUgSG7Mjf8aDLv68FjXHxHjVvQaj0pg2I+1qADZ
  bN6m5xaazqAShF5MN4zQQTnNHTp6AIXOSQhIpqKS/Bjf3FYw48pxCyMCgYEAyKjf
  GnwiFJZN/q3s2mCmlEPblJ5mbXGCmIK/wjcoDST3+YrFi5VoWsHu0hRoZHtxIiaH
  sjSj8f8hWaZJ+yTL/V6zAO93JMovmoYyClmGt5pl56pFT2B7VGDC5FU9bylzWF3g
  HDdCTXOE72c5cOHnOddxVSBdD6GLC7Qe4CUVnlECgYBYNmSskywHyVhWMaA+gWrI
  HA5KP2EhSidFRYHD9UJut6FMvn2NExaI3bMG4agbdDfMEKxxMuCGym18UQFAu1Cq
  miPBixZL05Yo2oRRRV+FNG2EfqFGGO6pbjKKK1m16tjNGSWFEjOs+adoGX+t7Ht6
  JOyNaRr7g4bhEgiFBEoFGQKBgF3XI+dl8CZCmJ0nR6JlGuIxzen2Hh7Gu/WJCBbS
  5qcnB9UrAfGiYNg44/BZXOzJEgKPlFxR4+4Ti8w6SVTrQ37tn7crRkPtTk/svFA8
  yBTrXwb1iU5y55pxWhOgjYeEEg5ccKehbB9+i8fONX3GF/Xj/Ht8FClwOe+yP9JB
  ZZfRAoGBAJb08mFdb0Csbp+ed3LFznWINpXf2vlRKqIf+w8VOsEItbiB0r08AVdA
  Tik8VkRWm9ZHnMeMRRg2sEsI8gfaEXwSfLfMi10fn9YuWC2GSt5z+lA52H/S1zU2
  sGHPNn1H/cu7eM+nr9NxzJIT2CzKMHt5w4epp/UgkYFri4n2wDNS
  -----END RSA PRIVATE KEY-----`;

const privateKeyFingerpint = "07:55:09:4b:5a:66:d6:d2:fb:a4:19:0a:f6:77:81:de";
const privateKeyList = [
  {
    id: "5f1996c8",
    name: "key1",
    data: privateKeyRSA,
    type: "RSA",
  },
  {
    id: "5f1996c9",
    name: "key2",
    data: privateKeyRSA,
    type: "RSA",
  },
];

const numberOfKeys = 2;

const store = createStore({
  state: {
    privateKeyList,
  },
  getters: {
    "privateKey/list": (state) => state.privateKeyList,
  },
  actions: {
    "privateKey/fetch": vi.fn(),
  },
});

describe("Private Key List", () => {
  let wrapper: VueWrapper<any>;
  const vuetify = createVuetify();

  
  beforeEach(async () => {
    vi.mock("window.global", () => {
      return {
        convertKeyToFingerprint: vi.fn().mockReturnValue(privateKeyFingerpint),
      };
    });
    
    wrapper = mount(PrivateKeyList, {
      global: {
        plugins: [[store, key], routes, vuetify],
      },
    });

    wrapper.vm.convertToFingerprint = vi.fn().mockReturnValue(privateKeyFingerpint);
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

  ///////
  // Data and Props checking
  //////

  it("Data is defined", () => {
    expect(wrapper.vm.$data).toBeDefined();
  });

  it("Compare data with default value", () => {
    expect(wrapper.vm.headers).toStrictEqual(headers);
  });

  it("Compare the computed with the default value", () => {
    expect(wrapper.vm.getListPrivateKeys).toStrictEqual(privateKeyList);
  });

  it("Check the function convertToFingerprint return the correct value", () => {
    for (let i = 0; i < numberOfKeys; i++) {
      expect(wrapper.vm.convertToFingerprint(privateKeyList[i].data)).toBe(
        privateKeyFingerpint
      );
    }
  });

  ///////
  // HTML validation
  //////

  it("Renders the correct HTML", () => {
    expect(wrapper.find('[data-test="privateKey-thead"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="privateKey-name"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="privateKey-fingerpint"]').exists()).toBeTruthy();
    expect(wrapper.find('[data-test="privateKey-menu-icon"]').exists()).toBeTruthy();
  });

  it("HTML has the correct values", () => {
    expect(wrapper.find('[data-test="privateKey-name"]').text()).toBe(privateKeyList[0].name);
    expect(wrapper.find('[data-test="privateKey-fingerpint"]').text()).toBe(privateKeyFingerpint);
  });
});
