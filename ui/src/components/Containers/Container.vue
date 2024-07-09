<template>
  <v-card class="bg-v-theme-surface">
    <v-tabs background-color="secondary" stacked color="primary">
      <v-tab to="/containers"> Container List </v-tab>
      <v-tab to="/containers/pending"> Pending </v-tab>
      <v-tab to="/containers/rejected"> Rejected </v-tab>
    </v-tabs>
    <v-divider />
  </v-card>

  <v-card>
    <router-view />
  </v-card>
</template>
