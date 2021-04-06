<template>
  <fragment>
    <v-btn
      color="red darken-1"
      outlined
      @click="dialog = !dialog"
    >
      Delete namespace
    </v-btn>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          This action cannot be undone. This will permanently delete the
          <b> {{ displayOnlyTenCharacters(name) }} </b> namespace and remove all sensors
          related to this namespace.
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            @click="remove();"
          >
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'NamespaceDelete',

  props: {
    nsTenant: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      name: '',
      dialog: false,
    };
  },

  computed: {
    tenant() {
      return this.$props.nsTenant;
    },
  },

  async created() {
    this.name = this.$store.getters['namespaces/get'].name;
  },

  methods: {
    async remove() {
      try {
        this.dialog = !this.dialog;
        await this.$store.dispatch('namespaces/remove', this.tenant);
        await this.$store.dispatch('auth/logout');
        await this.$router.push({ name: 'login' });
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceDelete);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceDelete);
      }
    },

    displayOnlyTenCharacters(str) {
      if (str !== undefined) {
        if (str.length > 10) return `${str.substr(0, 10)}...`;
      }
      return str;
    },
  },
};

</script>
