<template>
  <v-form
    v-show="show"
  >
    <v-row>
      <v-col
        md="auto"
        class="pr-0"
      >
        <h3 class="mb-8 ml-6">
          Security
        </h3>
        <b>
          <!-- @change="setSessionRecord" -->
          <v-checkbox
            v-model="sessionRecord"
            label="Enable session record"
            class="pt-0 mt-0 ml-6"
          />
        </b>
        <p class="ml-6">
          Session record is a feature that allows you to check logged activity when
          connecting to a device.
        </p>
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
