<template>
  <fragment>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="mt-4 mb-4"
      >
        <v-col
          sm="8"
        >
          <v-card class="pb-0 elevation-0">
            <div class="d-flex pa-0 align-center">
              <v-spacer />
              <v-spacer />
              <TokenAdd
                @update="getTokens"
              />
            </div>

            <TokenList />
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </fragment>
</template>

<script>

import TokenAdd from '@/components/setting/token/TokenAdd';
import TokenList from '@/components/setting/token/TokenList';

export default {
  name: 'SettingPrivateKeys',

  components: {
    TokenAdd,
    TokenList,
  },

  created() {
    this.getTokens();
  },

  methods: {
    async getTokens() {
      try {
        await this.$store.dispatch('tokens/fetch');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.tokenList);
      }
    },
  },
};
</script>
