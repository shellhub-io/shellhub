<template>
  <div class="d-flex pa-0 align-center">
    <h1>User Details</h1>
  </div>
  <v-card
    v-if="user.id"
    class="mt-2 border rounded bg-background"
    elevation="0"
  >
    <v-card-title
      class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface"
    >
      <span class="text-h6 ml-2 d-flex align-center ga-2">
        {{ user.username || user.name || user.email || 'User' }}
        <v-chip
          v-if="user.admin"
          color="warning"
          data-test="user-admin-chip"
        >
          <v-icon
            icon="mdi-crown"
            class="mt-n1 mr-1"
          />
          Admin
        </v-chip>
      </span>

      <v-menu
        location="bottom"
        scrim
        eager
      >
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            variant="plain"
            class="border rounded bg-v-theme-background"
            density="comfortable"
            size="default"
            icon="mdi-format-list-bulleted"
            data-test="user-actions-menu-btn"
          />
        </template>

        <v-list
          class="bg-v-theme-surface"
          lines="two"
          density="compact"
        >
          <v-list-item
            data-test="user-login-token-btn"
            @click="loginWithToken"
          >
            <div class="d-flex align-center">
              <v-icon class="mr-2">mdi-login</v-icon>
              <v-list-item-title>Login with token</v-list-item-title>
            </div>
          </v-list-item>

          <UserDelete
            :id="userId"
            v-slot="{ openDialog }"
            redirect
          >
            <div>
              <v-list-item
                data-test="user-delete-btn"
                @click="openDialog"
              >
                <div class="d-flex align-center">
                  <v-icon class="mr-2">mdi-delete</v-icon>
                  <v-list-item-title>Delete this user</v-list-item-title>
                </div>
              </v-list-item>
            </div>
          </UserDelete>

          <div class="px-2 py-1" />
        </v-list>
      </v-menu>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="user-uid-field">
            <div class="item-title">UID:</div>
            <p class="text-truncate">
              <code>{{ user.id }}</code>
            </p>
          </div>

          <div
            v-if="user.name"
            data-test="user-name-field"
          >
            <div class="item-title">Name:</div>
            <p class="text-truncate">{{ user.name }}</p>
          </div>

          <div
            v-if="user.username"
            data-test="user-username-field"
          >
            <div class="item-title">Username:</div>
            <p class="text-truncate">{{ user.username }}</p>
          </div>

          <div
            v-if="user.email"
            data-test="user-email-field"
          >
            <div class="item-title">Email:</div>
            <p class="text-truncate">{{ user.email }}</p>
          </div>

          <div
            v-if="user.recovery_email"
            data-test="user-recovery-email-field"
          >
            <div class="item-title">Recovery Email:</div>
            <p class="text-truncate">{{ user.recovery_email }}</p>
          </div>

          <div
            v-if="user.status"
            data-test="user-status-field"
          >
            <div class="item-title">Status:</div>
            <UserStatusChip
              :status="user.status"
            />
          </div>
        </v-col>

        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div
            v-if="user.created_at"
            data-test="user-created-field"
          >
            <div class="item-title">Created:</div>
            <p>{{ formatFullDateTime(user.created_at) }}</p>
          </div>

          <div
            v-if="user.last_login"
            data-test="user-last-login-field"
          >
            <div class="item-title">Last Login:</div>

            <template v-if="!lastLoginText">
              <p class="text-disabled">
                User never logged in
              </p>
            </template>

            <template v-else>
              <p>{{ lastLoginText }}</p>
            </template>
          </div>

          <div
            class="d-flex align-center flex-wrap mt-2"
            data-test="user-max-namespace-row"
          >
            <div class="mr-6">
              <div class="item-title">Max Namespaces:</div>
              <p>{{ user.max_namespaces }}</p>
            </div>
            <div>
              <div class="item-title">Namespaces Owned:</div>
              <p>{{ user.namespacesOwned }}</p>
            </div>
          </div>

          <div
            class="d-flex align-center flex-wrap mt-2"
            data-test="user-mfa-marketing-row"
          >
            <div class="mr-6">
              <div class="item-title">MFA:</div>
              <v-chip
                size="small"
                :color="user.mfa?.enabled ? 'success' : 'warning'"
              >
                {{ user.mfa?.enabled ? 'Enabled' : 'Disabled' }}
              </v-chip>
            </div>
            <div
              v-if="user.email_marketing !== undefined && user.email_marketing !== null"
            >
              <div class="item-title">Marketing Emails:</div>
              <v-chip
                size="small"
                :color="user.email_marketing ? 'primary' : 'default'"
              >
                {{ user.email_marketing ? 'Opted in' : 'Opted out' }}
              </v-chip>
            </div>
          </div>

          <div
            v-if="authMethods.length"
            class="mt-4"
            data-test="user-auth-methods-field"
          >
            <div class="item-title">Auth Methods:</div>
            <div class="d-flex flex-wrap">
              <v-chip
                v-for="(method, i) in authMethods"
                :key="`auth-${i}`"
                size="small"
                class="mr-2 mb-2"
              >
                {{ method }}
              </v-chip>
            </div>
          </div>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>

  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onBeforeMount, ref } from "vue";
import { useRoute } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import useAuthStore from "@admin/store/modules/auth";
import UserStatusChip from "@admin/components/User/UserStatusChip.vue";
import UserDelete from "@admin/components/User/UserDelete.vue";
import type { IAdminUser } from "@admin/interfaces/IUser";
import useSnackbar from "@/helpers/snackbar";
import { formatFullDateTime } from "@/utils/date";

const route = useRoute();
const snackbar = useSnackbar();
const usersStore = useUsersStore();
const authStore = useAuthStore();

const userId = computed(() => route.params.id as string);
const user = ref({} as IAdminUser);

const lastLoginText = computed(() => {
  const value = user.value?.last_login;
  if (!value || value === "0001-01-01T00:00:00Z") {
    return null;
  }
  return formatFullDateTime(value);
});

const authMethods = computed<string[]>(() => {
  const list = user.value?.preferences?.auth_methods;
  return Array.isArray(list) ? list : [];
});

const loginWithToken = async () => {
  try {
    const token = await authStore.getLoginToken(userId.value);
    const url = `/login?token=${token}`;
    window.open(url, "_blank", "noopener");
  } catch {
    snackbar.showError("Failed to get the login token.");
  }
};

onBeforeMount(async () => {
  try {
    user.value = await usersStore.fetchUserById(userId.value);
  } catch {
    snackbar.showError("Failed to get user details.");
  }
});

defineExpose({ user });
</script>

<style lang="scss" scoped>
.item-title {
  margin-top: 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1666666667em;
  line-height: 2.667;
}
</style>
