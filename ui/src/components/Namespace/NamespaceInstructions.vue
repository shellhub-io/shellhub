<template>
  <WindowDialog
    v-model="showNoNamespaceDialog"
    title="You have no namespaces associated"
    icon="mdi-folder-alert"
    icon-color="warning"
    @close="showNoNamespaceDialog = false"
  >
    <div class="pa-6">
      <p>
        In order to use ShellHub, you must have a namespace associated
        with your account or join an existing one.
      </p>

      <div v-if="isCommunity" id="cli-instructions">
        <p class="mt-3">
          The easiest way to configure a namespace is by using the <strong>CLI script</strong>.
          Once you add a namespace via CLI script, this dialog will be
          automatically closed.
        </p>
      </div>
    </div>
    <template #footer>
      <div v-if="isCommunity" class="d-flex align-center w-100 px-6">
        <span class="text-caption text-center">
          For more information, check out the
          <a
            :href="'https://docs.shellhub.io/self-hosted/administration'"
            target="_blank"
            rel="noopener noreferrer"
            class="text-decoration-none text-primary"
            data-test="openContentSecond-text"
          >
            ShellHub Administration Guide
            <v-icon size="12" class="ml-1">mdi-open-in-new</v-icon>
          </a>
        </span>
      </div>

      <v-card-actions v-else class="d-flex justify-end w-100">
        <v-btn
          color="primary"
          @click="showNamespaceAdd = true"
          data-test="save-btn"
        >
          Add Namespace
        </v-btn>
      </v-card-actions>
    </template>

    <NamespaceAdd
      v-model="showNamespaceAdd"
      enableSwitchIn
      data-test="namespace-add-component"
    />
  </WindowDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRoute } from "vue-router";
import { envVariables } from "@/envVariables";
import NamespaceAdd from "./NamespaceAdd.vue";
import WindowDialog from "../WindowDialog.vue";

const route = useRoute();
const showDialog = defineModel<boolean>({ default: false });
const showNamespaceAdd = ref(false);
const showNoNamespaceDialog = ref(route.name === "AcceptInvite" ? false : showDialog.value);
const { isCommunity } = envVariables;
</script>
