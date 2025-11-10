import { createVuetify } from "vuetify";
import { flushPromises, mount } from "@vue/test-utils";
import { describe, expect, it, vi, beforeAll } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { createMemoryHistory, createRouter } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import useAuthStore from "@admin/store/modules/auth";
import UserDetails from "@admin/views/UserDetails.vue";
import { SnackbarPlugin } from "@/plugins/snackbar";

vi.mock("@/utils/date", () => ({
  formatFullDateTime: (iso: string) => `formatted(${iso})`,
}));

const mockUser = {
  id: "6256b739302b50b6cc5eafcc",
  status: "confirmed",
  created_at: "2022-04-13T11:42:49.578Z",
  last_login: "0001-01-01T00:00:00Z",
  email: "antony@gmail.com",
  name: "antony",
  username: "antony",
  namespacesOwned: 2,
  max_namespaces: 0,
  mfa: { enabled: false },
  email_marketing: null,
  preferences: { auth_methods: ["local"] },
};

describe("UserDetails.vue", async () => {
  const pinia = createPinia();
  setActivePinia(pinia);

  const usersStore = useUsersStore();
  usersStore.fetchUserById = vi.fn().mockResolvedValue(mockUser);

  const authStore = useAuthStore();
  authStore.getLoginToken = vi.fn().mockResolvedValue("mock-token");

  const vuetify = createVuetify();

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: "/user/:id",
        name: "userDetails",
        component: UserDetails,
      },
    ],
  });

  await router.push({ name: "userDetails", params: { id: mockUser.id } });
  await router.isReady();

  const wrapper = mount(UserDetails, {
    global: {
      plugins: [pinia, vuetify, router, SnackbarPlugin],
      stubs: { Teleport: true },
    },
  });

  beforeAll(async () => {
    await flushPromises();
  });

  it("is a Vue instance", () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("loads and exposes the user data", () => {
    expect((wrapper.vm).user).toEqual(mockUser);
  });

  it("renders the title", () => {
    expect(wrapper.find("h1").text()).toBe("User Details");
  });

  it("renders main user fields in their data-test blocks", () => {
    const uid = wrapper.find("[data-test='user-uid-field']");
    expect(uid.exists()).toBe(true);
    expect(uid.text()).toContain(mockUser.id);

    const email = wrapper.find("[data-test='user-email-field']");
    expect(email.exists()).toBe(true);
    expect(email.text()).toContain(mockUser.email);

    const username = wrapper.find("[data-test='user-username-field']");
    expect(username.exists()).toBe(true);
    expect(username.text()).toContain(mockUser.username);

    const status = wrapper.find("[data-test='user-status-field']");
    expect(status.exists()).toBe(true);
    expect(status.text().toLowerCase()).toContain("confirmed");
  });

  it("renders created_at using the formatter", () => {
    const created = wrapper.find("[data-test='user-created-field']");
    expect(created.exists()).toBe(true);
    expect(created.text()).toContain(`formatted(${mockUser.created_at})`);
  });

  it("renders 'never logged in' state for the zero-date sentinel", () => {
    const lastLogin = wrapper.find("[data-test='user-last-login-field']");
    expect(lastLogin.exists()).toBe(true);
    expect(lastLogin.text()).toContain("User never logged in");
  });

  it("renders MFA and Marketing blocks", () => {
    const row = wrapper.find("[data-test='user-mfa-marketing-row']");
    expect(row.exists()).toBe(true);

    expect(row.text()).toContain("Disabled");

    expect(row.text()).not.toContain("Opted in");
    expect(row.text()).not.toContain("Opted out");
  });

  it("renders auth methods chips", () => {
    const auth = wrapper.find("[data-test='user-auth-methods-field']");
    expect(auth.exists()).toBe(true);
    expect(auth.text()).toContain("local");
  });

  it("renders namespace counters", () => {
    const row = wrapper.find("[data-test='user-max-namespace-row']");
    expect(row.exists()).toBe(true);
    expect(row.text()).toContain(String(mockUser.max_namespaces));
    expect(row.text()).toContain(String(mockUser.namespacesOwned));
  });
});
