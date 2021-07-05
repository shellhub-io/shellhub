<template>
  <v-card>
    <v-dialog
      v-model="showNoNamespace"
      :retain-focus="false"
      max-width="650px"
      persistent
    >
      <v-card>
        <v-card-title
          class="headline grey lighten-2"
        >
          There are no namespaces associated with your account
        </v-card-title>

        <v-card-text class="mt-4 mb-0 pb-1 mb-4">
          <p>
            In order to use ShellHub, you first need to create a namespace to associate with
            your account or join an existing one.
          </p>
          <div
            v-if="openVersion"
            id="cli-instructions"
          >
            <p data-test="openContentFirst-text">
              The easiest way to configure a namespace is by using the cli
              script.
            </p>
            <p
              class="caption mb-0"
              data-test="openContentSecond-text"
            >
              Check the <a
                :href="'https://docs.shellhub.io/admin-manual/managing/'"
                target="_blank"
              >documentation</a>
              for more information and alternative install methods.
            </p>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-btn
            text
            @click="close"
          >
            Close
          </v-btn>
          <v-spacer />
          <v-btn
            v-if="!openVersion"
            id="namespace-add"
            text
            @click="dialogAdd = !dialogAdd"
          >
            Add Namespace
          </v-btn>
        </v-card-actions>
      </v-card>
      <NamespaceAdd
        v-if="!openVersion"
        :show.sync="dialogAdd"
        :first-namespace="autoSwitch"
        data-test="namespace-btn"
      />
    </v-dialog>
  </v-card>
</template>

<script>

import NamespaceAdd from '@/components/app_bar/namespace/NamespaceAdd';

export default {
  name: 'NamespaceInstructions',

  components: {
    NamespaceAdd,
  },

  props: {
    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      dialogAdd: false,
    };
  },

  computed: {
    showNoNamespace: {
      get() {
        return this.$props.show;
      },

      set(value) {
        this.$emit('show', value);
      },
    },

    openVersion() {
      return !this.$env.isEnterprise;
    },

    autoSwitch() {
      return localStorage.getItem('tenant') === '';
    },
  },

  methods: {
    close() {
      this.$emit('update:show', false);
    },
  },
};
</script>
