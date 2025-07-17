<template>
  <BaseDialog v-model="showNoNamespaceDialog" :retain-focus="false">
    <v-card
      class="bg-v-theme-surface"
    >
      <v-card-title class="bg-primary">
        There are no namespaces associated with your account
      </v-card-title>

      <v-card-text class="mt-4 mb-0 pb-1 mb-4">
        <p class="text-body-2">
          In order to use ShellHub, you must have a namespace associate
          with your account or join an existing one.
        </p>
        <div v-if="openVersion" id="cli-instructions" class="mt-3 text-body-2">
          <p data-test="openContentFirst-text">
            The easiest way to configure a namespace is by using the cli script.
          </p>
          <p class="mt-3" data-test="cliUpdateWarning-text">
            When you add a namespace, on cli script, this dialog will be
            automatically closed.
          </p>
          <p class="text-caption mb-0 mt-3" data-test="openContentSecond-text">
            Check the
            <a
              :href="'https://docs.shellhub.io/self-hosted/administration'"
              target="_blank"
              rel="noopener noreferrer"
            >documentation</a
            >
            for more information and alternative install methods.
          </p>
        </div>
      </v-card-text>

      <v-card-actions v-if="!openVersion">
        <v-spacer />
        <div>
          <v-btn
            color="primary"
            @click="showNamespaceAdd = true"
            data-test="save-btn">
            Add Namespace
          </v-btn>
          <NamespaceAdd
            v-model="showNamespaceAdd"
            enableSwitchIn
            data-test="namespace-add-component"
          />
        </div>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { envVariables } from "@/envVariables";
import NamespaceAdd from "./NamespaceAdd.vue";
import BaseDialog from "../BaseDialog.vue";

const route = useRoute();

const showDialog = defineModel<boolean>({ default: false });

const showNamespaceAdd = ref(false);

const showNoNamespaceDialog = computed(() => route.name === "AcceptInvite" ? false : showDialog.value);

const openVersion = computed(() => !envVariables.isCloud && !envVariables.isEnterprise);
</script>
