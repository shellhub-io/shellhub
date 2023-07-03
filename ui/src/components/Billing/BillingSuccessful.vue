<template>
  <v-row>
    <v-col align="center">
      <v-card-title data-test="title">Congratulations! Subscription Successful!</v-card-title>
    </v-col>
  </v-row>
  <v-row>
    <v-col align="center">
      <v-icon
        end
        icon="mdi-cloud-check-variant"
        color="green"
        size="100"
        class="green-cloud"
        data-test="green-cloud-icon"
      />
      <v-spacer class="mt-4" />
      <h4 data-test="h4-letter">Thank you for choosing ShellHub Cloud! Your subscription has been successfully activated,
        and we're thrilled to have you on board. </h4>
    </v-col>
  </v-row>
  <v-row>
    <v-col>
      <p data-test="p-letter">You now have access to our powerful device management solution,
        allowing you to effortlessly connect, monitor, and control your devices from anywhere.</p>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
</script>

<style scoped>
.green-cloud {
  filter: drop-shadow(0px 0px 30px rgba(43, 255, 10, 0.444))
}
</style>
