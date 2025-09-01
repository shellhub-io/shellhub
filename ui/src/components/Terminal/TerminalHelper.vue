<template>
  <BaseDialog v-model="showDialog" @close="close" transition="dialog-bottom-transition" data-test="dialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h6 pa-4 bg-primary" data-test="dialog-title">
        Generate a SSH command line
      </v-card-title>

      <v-card-text class="pt-4 pb-0">
        <v-row class="mt-1">
          <v-col>
            <v-text-field
              v-model="username"
              variant="outlined"
              label="Username"
              hint="Enter an existing user on the device"
              persistent-placeholder
              persistent-hint
              density="compact"
              data-test="username-input"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <CopyWarning :copied-item="'Command'">
              <template #default="{ copyText }">
                <v-text-field
                  :model-value="commandLine"
                  @click:append-inner="copyText(commandLine)"
                  append-inner-icon="mdi-content-copy"
                  hint="Run this command on your Terminal"
                  class="code"
                  variant="outlined"
                  label="Command Line"
                  readonly
                  density="compact"
                  persistent-placeholder
                  persistent-hint
                  data-test="command-field"
                />
              </template>
            </CopyWarning>
          </v-col>
        </v-row>

        <v-row v-if="showCheckbox">
          <v-col class="d-flex justify-center align-center">
            <v-checkbox
              v-model="dispenseHelper"
              label="Always copy SSHID directly and skip this helper"
              density="compact"
              hide-details
              data-test="dispense-checkbox"
            />
          </v-col>
        </v-row>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="showDialog = false">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import BaseDialog from "../BaseDialog.vue";

interface Props {
  sshid: string;
  userId?: string;
  showCheckbox?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  userId: "",
  showCheckbox: false,
});

const showDialog = defineModel({ default: false });
const username = ref("");
const dispenseHelper = ref(false);
const LS_KEY = "dispenseTerminalHelper";

const getDispensedUsers = (): string[] => {
  try {
    return JSON.parse(localStorage.getItem(LS_KEY) || "[]");
  } catch {
    return [];
  }
};

const setDispensedUsers = (users: string[]) => {
  localStorage.setItem(LS_KEY, JSON.stringify(users));
};

const commandLine = computed(() => {
  const trimmedUsername = username.value.trim();
  return trimmedUsername ? `ssh ${trimmedUsername}@${props.sshid}` : "";
});

watch(dispenseHelper, (isDispensed) => {
  const users = getDispensedUsers();
  const userIndex = users.indexOf(props.userId);

  if (isDispensed && userIndex === -1) {
    users.push(props.userId);
  } else if (!isDispensed && userIndex !== -1) {
    users.splice(userIndex, 1);
  }

  setDispensedUsers(users);
});

watch(showDialog, (isOpen) => {
  if (isOpen) {
    const users = getDispensedUsers();
    dispenseHelper.value = users.includes(props.userId);
  }
});

const close = () => {
  username.value = "";
  showDialog.value = false;
};

defineExpose({ showDialog });
</script>

<style scoped lang="scss">
.code ::v-deep(.v-field__input) {
  font-family: monospace;
  font-size: 85%;
}
</style>
