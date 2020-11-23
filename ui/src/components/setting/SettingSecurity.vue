<template>
  <v-form
    v-show="show"
  >
    <v-row>
      <v-col>
        <h3 class="mb-5">
          Security
        </h3>

        <div
          class="ml-3"
        >
          <v-checkbox
            v-model="sessionRecord"
            label="Enable session record"
          />

          Session record is a feature that allows you to check logged activity when
          connecting to a device.
        </div>
      </v-col>
    </v-row>
  </v-form>
</template>

<script>

export default {
  name: 'SettingSecurity',

  props: {
    show: {
      type: Boolean,
      required: true,
    },
  },

  computed: {
    sessionRecord: {
      get() {
        return this.$store.getters['security/get'];
      },

      async set(value) {
        try {
          await this.$store.dispatch('security/set', value);
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      },
    },
  },

  async created() {
    if (this.show) {
      try {
        await this.$store.dispatch('security/get');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorDefault');
      }
    }
  },
};
</script>
