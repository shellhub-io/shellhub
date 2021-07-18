<template>
  <fragment>
    <v-btn
      class="v-btn--active mr-2"
      text
      color="primary"
      @click="createToken()"
    >
      Add Token
    </v-btn>
  </fragment>
</template>

<script>

export default {
  name: 'DeviceAdd',

  methods: {
    async createToken() {
      try {
        await this.$store.dispatch('tokens/post');
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.tokenCreating);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.tokenCreating);
      }
    },
  },
};

</script>
