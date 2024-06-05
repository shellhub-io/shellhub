<template>
  <v-dialog v-model="showNoNamespace" :retain-focus="false" persistent max-width="650px">
    <v-card
      v-model="showNoNamespace"

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

      <v-card-actions>
        <v-spacer />
        <div>
          <NamespaceAdd
            v-if="!openVersion"
            enableSwitchIn
            data-test="namespaceAdd-component"
          />
        </div>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { envVariables } from "../../envVariables";
import NamespaceAdd from "./NamespaceAdd.vue";

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(["update"]);

const showNoNamespace = computed({
  get() {
    return props.show;
  },
  set(value: boolean) {
    emit("update", value);
  },
});

const openVersion = computed(() => !envVariables.isCloud || !envVariables.isEnterprise);
</script>
