<template>
  <WindowDialog
    v-model="showDialog"
    title="Account Deletion"
    :description="dialogDescription"
    :icon="dialogIcon"
    icon-color="primary"
    @close="showDialog = false"
  >
    <v-card-text class="pa-6">
      <div v-if="isCommunity">
        In Community instances, user accounts can only be deleted via the CLI.
        For detailed instructions on user management, please refer to our
        <a
          href="https://docs.shellhub.io/self-hosted/administration#delete-a-user"
          target="_blank"
          rel="noopener noreferrer"
          class="text-primary"
          data-test="docs-link"
        >
          administration documentation
          <v-icon size="16" class="ml-1" icon="mdi-open-in-new" />
        </a>.

        <div>
          <p class="text-subtitle-2 mt-4 mb-2">Run this command to delete your account:</p>
          <CopyWarning copied-item="Command">
            <template #default="{ copyText }">
              <v-text-field
                :model-value="'$ ' + deleteCommand"
                class="code"
                readonly
                density="compact"
                hide-details
              >
                <template #append>
                  <v-btn
                    icon="mdi-content-copy"
                    color="primary"
                    variant="flat"
                    rounded
                    size="small"
                    @click="copyText(deleteCommand)"
                  />
                </template>
              </v-text-field>
            </template>
          </CopyWarning>
        </div>
      </div>

      <div v-else>
        <span>
          In Enterprise instances, user accounts can only be deleted via the Admin Console.
          Please access your
          <a
            class="font-weight-medium text-primary"
            href="/admin/users"
            target="_blank"
            rel="noopener noreferrer"
          >Admin Console</a>
          or contact your system administrator for assistance.
        </span>
      </div>
    </v-card-text>

    <template #footer>
      <v-spacer />
      <v-btn
        variant="text"
        @click="showDialog = false"
        data-test="close-btn"
      >
        Close
      </v-btn>
    </template>
  </WindowDialog>
</template>

<script setup lang="ts">
import WindowDialog from "../WindowDialog.vue";
import CopyWarning from "../User/CopyWarning.vue";
import useAuthStore from "@/store/modules/auth";
import { envVariables } from "@/envVariables";

const showDialog = defineModel<boolean>({ default: false });
const { username } = useAuthStore();
const { isCommunity } = envVariables;
const deleteCommand = `./bin/cli user delete ${username}`;
const dialogDescription = isCommunity ? "CLI Required" : "Admin Console Required";
const dialogIcon = isCommunity ? "mdi-console" : "mdi-shield-account";

defineExpose({ showDialog });
</script>

<style scoped lang="scss">
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
