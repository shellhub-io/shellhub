<template>
  <v-dialog v-model="showNoNamespace" max-width="650px">
    <v-card
      v-model="showNoNamespace"
      :retain-focus="false"
      persistent
      class="bg-v-theme-surface"
    >
      <v-card-title class="bg-primary">
        There are no namespaces associated with your account
      </v-card-title>

      <v-card-text class="mt-4 mb-0 pb-1 mb-4">
        <p class="text-body-2">
          In order to use ShellHub, you first need to create a namespace to
          associate with your account or join an existing one.
        </p>
        <div v-if="openVersion" id="cli-instructions" class="mt-3 text-body-2">
          <p data-test="openContentFirst-text">
            The easiest way to configure a namespace is by using the cli script.
          </p>
          <p class="text-caption mb-0 mt-3" data-test="openContentSecond-text">
            Check the
            <a
              :href="'https://docs.shellhub.io/admin-manual/managing/'"
              target="_blank"
              >documentation</a
            >
            for more information and alternative install methods.
          </p>
        </div>
      </v-card-text>

      <v-card-actions>
        <v-btn text @click="close"> Close </v-btn>
        <v-spacer />
        <v-btn
          v-if="!openVersion"
          id="namespace-add"
          text
          data-test="add-btn"
          @click="dialogAdd = !dialogAdd"
        >
          Add Namespace
        </v-btn>
      </v-card-actions>
    </v-card>

    <NamespaceAdd
      v-if="!openVersion"
      :show="dialogAdd"
      :firstNamespace="autoSwitch"
      data-test="namespaceAdd-component"
    />
  </v-dialog>
</template>

<script lang="ts">
import { envVariables } from "../../envVariables";
import { defineComponent, ref, computed } from "vue";
import NamespaceAdd from "./NamespaceAdd.vue";

export default defineComponent({
  props: {
    show: {
      type: Boolean,
      required: true,
    },
  },
  setup(props, ctx) {
    const dialogAdd = ref(false);

    const showNoNamespace = computed({
      get() {
        return props.show;
      },
      set(value: boolean) {
        ctx.emit("update", value);
      },
    });

    const openVersion = computed(() => envVariables.isEnterprise);

    const autoSwitch = computed(() => localStorage.getItem("tenant") === "");

    const close = () => {
      showNoNamespace.value = false;
      ctx.emit("update", false);
    };

    return {
      dialogAdd,
      openVersion,
      showNoNamespace,
      autoSwitch,
      close,
    };
  },
  components: { NamespaceAdd },
});
</script>
