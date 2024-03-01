import { createVuetify } from "vuetify";
import { DOMWrapper, flushPromises, mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MockAdapter from "axios-mock-adapter";
import { nextTick } from "vue";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
import { namespacesApi, usersApi } from "@/api/http";
import { store, key } from "@/store";
import { router } from "@/router";
import { envVariables } from "@/envVariables";
import { SnackbarPlugin } from "@/plugins/snackbar";
import { INotificationsSuccess } from "@/interfaces/INotifications";

type PrivateKeyAddWrapper = VueWrapper<InstanceType<typeof PrivateKeyAdd>>;

describe("Setting Private Keys", () => {
  const node = document.createElement("div");
  node.setAttribute("id", "app");
  document.body.appendChild(node);

  let wrapper: PrivateKeyAddWrapper;

  const vuetify = createVuetify();

  let mockNamespace: MockAdapter;

  let mockUser: MockAdapter;

  const members = [
    {
      id: "xxxxxxxx",
      username: "test",
      role: "owner",
    },
  ];

  const namespaceData = {
    name: "test",
    owner: "test",
    tenant_id: "fake-tenant-data",
    members,
    settings: {
      session_record: true,
    },
    max_devices: 3,
    devices_count: 3,
    created_at: "",
  };

  const authData = {
    status: "success",
    token: "",
    user: "test",
    name: "test",
    tenant: "fake-tenant-data",
    email: "test@test.com",
    id: "xxxxxxxx",
    role: "owner",
    mfa: {
      enable: false,
      validate: false,
    },
  };

  const session = true;

  const privateKeys = [{
    name: "test",
    data: `-----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAACFwAAAAdzc2gtcn
    NhAAAAAwEAAQAAAgEAneJ7kCYsKAwdSMgu37VNdtBu3Syr8S0kEnSaIjnYizKY/6BU6wC0
    zH9JsiuCAcC0Qc2m9jNAV/udlDO7tbra3WxXcKKxyPIQ0TOyvE8N1mR6IU8D6atyN7FM1b
    uatb95IB22qztLtMRHWoGb50NCbAIVv7wuwBtegYYDcyjT6v8YEBQvHkl01zpudOlIiFqE
    y1Wd+W4l48cTfbfc0iK84nakK6vi/lRWkliKF4t3L4s3NaoLJZ1ibDd1/yLzVnIvLpwM9k
    CZn4jvPp6fTFbFvfN2KvxEXDN18vqNkJlZh2lgKSS9ZJuHp5LZaUiBYcTiovo3LaV2q+yp
    NT61VbzzSy7+cnmGl5lpRuikrg6MeSwawpo+6sdSYvBdgxh/cDtKIPputkm6kMp/iizepW
    SzrPlAkCji7bP9v+q+sZLN0oe1zOm7rWX9qqqq3PBaHNCm26gXiHYsIXb2ntNC19K4PULt
    ZFg6EUkEgFlka4o2keFLAarY0Ogx7KX4I6gnnuEc8yZOdCrs3gsBuaNIAMoxKbwLIXdU7Y
    8vqFcGqc0R2ahKoFOmohMW/hySVDUVplmFj1Wq19Bpnc0LtWqRXeuQG92XOMMSQyxXx0c8
    rD8llhBoiIC2w1G58RZZOLRYNXE6/rN0UafN9BQypqRu6PmCzQmG+5U7qz4GZ7zQv8vyKI
    UAAAdI80sWO/NLFjsAAAAHc3NoLXJzYQAAAgEAneJ7kCYsKAwdSMgu37VNdtBu3Syr8S0k
    EnSaIjnYizKY/6BU6wC0zH9JsiuCAcC0Qc2m9jNAV/udlDO7tbra3WxXcKKxyPIQ0TOyvE
    8N1mR6IU8D6atyN7FM1buatb95IB22qztLtMRHWoGb50NCbAIVv7wuwBtegYYDcyjT6v8Y
    EBQvHkl01zpudOlIiFqEy1Wd+W4l48cTfbfc0iK84nakK6vi/lRWkliKF4t3L4s3NaoLJZ
    1ibDd1/yLzVnIvLpwM9kCZn4jvPp6fTFbFvfN2KvxEXDN18vqNkJlZh2lgKSS9ZJuHp5LZ
    aUiBYcTiovo3LaV2q+ypNT61VbzzSy7+cnmGl5lpRuikrg6MeSwawpo+6sdSYvBdgxh/cD
    tKIPputkm6kMp/iizepWSzrPlAkCji7bP9v+q+sZLN0oe1zOm7rWX9qqqq3PBaHNCm26gX
    iHYsIXb2ntNC19K4PULtZFg6EUkEgFlka4o2keFLAarY0Ogx7KX4I6gnnuEc8yZOdCrs3g
    sBuaNIAMoxKbwLIXdU7Y8vqFcGqc0R2ahKoFOmohMW/hySVDUVplmFj1Wq19Bpnc0LtWqR
    XeuQG92XOMMSQyxXx0c8rD8llhBoiIC2w1G58RZZOLRYNXE6/rN0UafN9BQypqRu6PmCzQ
    mG+5U7qz4GZ7zQv8vyKIUAAAADAQABAAACABunassh1IQjMxHndkZaxDm2YmS9CVTR+kp9
    P+4UwbgH4cKMe7M5yXE0Ll1Vv4y9CxWnhsIC0hdXDA/ES/GVy/YSnvIsnQU8WPO7oWfYVO
    0jZjzlUSMhk3zrwjCBjqSc6ANXEQLG/QiphHH2167XGhA/AT43IN0nLhNzvLD0CsJTcgyG
    7IXaieuU4Xn6zmiLqkzPLz9cKqjN2r0fcj8gNINaEoFPtw+jCBLUDUP4eqTKNp5grVkmSv
    H3eOR6Y7LVhywbyy1qvT2zR2xpbi9512Lg/OakjvizTsqDVj5ojcTpER3DwKSZlVYlo40M
    VRUh3ix7tSR6oeGVL0ITPMHKubZAER6f7+ead9R9WoNC8UFfT8yC1aZ3d/Lrkb45AVrYlo
    q04d20dOdAEzss61tLtY+vSNvYDD9UozxURGd8oOPhm7xOMMnILr6Spv1MyIEYQMvK55p6
    vmAFTQDGi7Qhxfeaj9MuvzFZeTdP9TqqDTJ26DxiWuPFYXLLF7hN+hNsT7vz95tU1Jf9tV
    8ZM+ahfFVRXqu/c4nZ9qkMxOGk5JHsxghE35fXYYi6F6WZH989HnotqOvDgQsck+jLsijH
    3Mvm2AUPN6ho4bibz6to0m2/1Q2GPrD8lYjdwbTvbxJnPjR5KH9oc1SWAGMmGiXYI1H/dF
    /FBhNXyz1LN2tW/WuBAAABAQCvoFTpu4X/4/h1VJoIqXT65bhUH0bMtiqELVjB1htInJWc
    ORkp5PCv7QdKX0drSba/7iinPTdB8Qw2SvS8TxlAy8A8kB/J8DSM5VKcfoTCG4AuND/Yze
    8ZXqLWPCgNMvHrtOoq8lHgMC25Rr2R8Zyz3XkZiRuYV9bJLeKxys7lubc4IGTJT5emu9cI
    dNjZ5otAMQD3ii3LbzKhJuvtAvIPB2UEJwBXRm08UhpAXYJmeXhAmytngEcWoTkvZVwe9o
    F/nwIIrydm3XIIo35sbwA3Fg/PNp4ENqkxxAD28vRlA1w9SEMOBbKfvSG7RXtXzbfEG+vH
    JpdL7Pp9Y0OICnvAAAABAQDM72yuqzrH0VuBF0WX6HtORk9CebpTcsSBab3hNg+OVo83+w
    SUoeyD8Zs4wkTv4XRfyeXJt2aETrDXRx3EfUM1yh6iUlPkdfy34nZblz873K7rypawrfp5
    g+269ur/B1LwBTptfC9qX0GXH0lReh15tnfA2z8+yfFa79mQz3/Mm+6OYRsbL2aLse9QKQ
    JDDKjU5pKYbIiiprTBEARTiOt3MN/n2M4vBnkrxV0/86FP2D4PyfsS897pv0nfDeLAGsW+
    5KYSo3dLdDhNigxl2UsAHzv+GmL8wuJfeEbNxHe+djnSpzOMgcZGtkcsHrKu4sz6jX99Gv
    WyN1orIrHZhBEhAAABAQDFOcJlTlCzXXfPra/iorxapek8vLU70gGY15X26U5BeK5XPEL5
    CSBXAJSP1L1AddEZ6Jn7b22QdcoE40b55+jOIHLB9OnjHCtAjuiD59JvkiE8DryYwk+s5O
    WDgw/UbMDUwm9lXqy1zHTE+pDJYyt/lVt9wTfTLBNd41lf4gDUmz0cw+el67O0csN3Hajs
    /ql/mMTh6YCn7IFqE1EoTsWeFuVmNZp3AFkmDGtr7El7SaUft9cWmo1GeSYVlXIR+iAXVv
    uGEOo9v/RHl0zPHBvph9MX+2unq4jDti1l8PzTvTcTlSZ4i98Xy5ttlGOlhcr2op1y1PFF
    0DuoApwWlhblAAAAEU15IHdlYi1zZXJ2ZXIga2V5AQ==
    -----END OPENSSH PRIVATE KEY-----`,
  }];

  beforeEach(async () => {
    const el = document.createElement("div");
    document.body.appendChild(el);
    vi.useFakeTimers();
    localStorage.setItem("tenant", "fake-tenant-data");
    envVariables.isCloud = true;

    mockNamespace = new MockAdapter(namespacesApi.getAxios());
    mockUser = new MockAdapter(usersApi.getAxios());

    mockNamespace.onGet("http://localhost:3000/api/namespaces/fake-tenant-data").reply(200, namespaceData);
    mockUser.onGet("http://localhost:3000/api/users/security").reply(200, session);
    mockUser.onGet("http://localhost:3000/api/auth/user").reply(200, authData);

    store.commit("auth/authSuccess", authData);
    store.commit("auth/changeData", authData);
    store.commit("namespaces/setNamespace", namespaceData);
    store.commit("security/setSecurity", session);
    wrapper = mount(PrivateKeyAdd, {
      global: {
        plugins: [[store, key], vuetify, router, SnackbarPlugin],
      },
      attachTo: el,
    });
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

  it("Renders components", async () => {
    const dialog = new DOMWrapper(document.body);

    await wrapper.findComponent('[data-test="private-key-dialog-btn"]').trigger("click");

    expect(wrapper.findComponent('[data-test="private-key-dialog-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="card-title"]').exists()).toBe(true);
    expect(dialog.find('[data-test="name-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-field"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-cancel-btn"]').exists()).toBe(true);
    expect(dialog.find('[data-test="private-key-save-btn"]').exists()).toBe(true);
  });

  it("Sets private key data error message", async () => {
    await wrapper.findComponent('[data-test="private-key-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="name-field"]').setValue("not-working-name");

    await wrapper.findComponent('[data-test="name-field"]').setValue("");

    await flushPromises();

    expect(wrapper.vm.nameError).toEqual("this is a required field");
  });

  it("Sets private key data error message", async () => {
    await wrapper.findComponent('[data-test="private-key-dialog-btn"]').trigger("click");

    await wrapper.findComponent('[data-test="private-key-field"]').setValue("not-working-key");

    await wrapper.findComponent('[data-test="private-key-field"]').setValue("");

    await flushPromises();

    expect(wrapper.vm.privateKeyDataError).toEqual("this is a required field");
  });
});
