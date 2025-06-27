<template>
  <v-dialog v-model="dialog" width="500" transition="dialog-bottom-transition" data-test="dialog">
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
                  @click:append="copyText(commandLine)"
                  class="code"
                  variant="outlined"
                  label="Command Line"
                  readonly
                  density="compact"
                  hide-details
                  data-test="command-field"
                />
              </template>
            </CopyWarning>
          </v-col>
        </v-row>

        <v-row>
          <v-col class="d-flex justify-center align-center">
            <CopyWarning :copied-item="'Command'">
              <template #default="{ copyText }">
                <v-btn color="primary" data-test="copy-btn" prepend-icon="mdi-content-copy" @click="copyText(commandLine)">
                  Copy
                </v-btn>
              </template>
            </CopyWarning>
          </v-col>
        </v-row>
        <v-row>
          <v-col class="d-flex justify-center align-center">
            <p class="text-body-2 align-center">
              Just copy and run on your terminal!
            </p>
          </v-col>
        </v-row>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" color="primary" data-test="close-btn" @click="dialog = false">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import CopyWarning from "@/components/User/CopyWarning.vue";

const props = defineProps({
  sshid: {
    type: String,
    required: true,
  },
});

const dialog = defineModel({ default: false });
const username = ref("");

const commandLine = computed(() => `ssh ${username.value}@${props.sshid}`);

defineExpose({ dialog });
</script>

<style scoped lang="scss">
.code {
  font-family: monospace;
  font-size: 85%;
}
</style>
