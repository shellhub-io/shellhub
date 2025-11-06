<template>
  <div class="pa-6">
    <div class="text-center mb-6">
      <v-avatar
        size="64"
        color="success"
        class="mb-4"
      >
        <v-icon
          size="32"
          color="white"
          icon="mdi-check-circle"
        />
      </v-avatar>
      <h2 class="text-h4 mb-2">
        All Set!
      </h2>
      <p class="text-subtitle-1 text-medium-emphasis mb-2">
        Your device has been successfully added
      </p>
    </div>

    <v-card
      variant="tonal"
      color="primary"
      class="pa-4 mb-6"
    >
      <div class="d-flex align-start">
        <v-icon
          size="24"
          color="primary"
          class="mr-3 mt-1"
        >
          mdi-party-popper
        </v-icon>
        <div>
          <h4 class="text-h6 mb-2">
            Welcome to ShellHub!
          </h4>
          <p class="text-body-1 mb-0">
            You're now ready to take advantage of all the powerful tools ShellHub provides.
            Connect to your devices, monitor sessions, manage access, and much more!
          </p>
        </div>
      </div>
    </v-card>

    <v-card
      variant="outlined"
      color="primary"
      class="pa-4 mb-4"
    >
      <div class="mb-3">
        <h4 class="text-h6 mb-3 d-flex align-center text-white">
          <v-icon
            color="primary"
            class="mr-2"
          >
            mdi-compass
          </v-icon>
          Helpful Resources
        </h4>
        <div class="d-flex flex-wrap ga-4">
          <v-btn
            href="http://docs.shellhub.io/"
            target="_blank"
            rel="noopener noreferrer"
            color="primary"
            variant="tonal"
            prepend-icon="mdi-book-open-variant"
            data-test="welcome-fourth-links"
          >
            Documentation
            <v-icon
              size="12"
              class="ml-1"
            >
              mdi-open-in-new
            </v-icon>
          </v-btn>
          <v-btn
            href="https://gitter.im/shellhub-io/community"
            target="_blank"
            rel="noopener noreferrer"
            color="primary"
            variant="tonal"
            prepend-icon="mdi-forum"
          >
            Community Chat
            <v-icon
              size="12"
              class="ml-1"
            >
              mdi-open-in-new
            </v-icon>
          </v-btn>
        </div>
      </div>
    </v-card>

    <div class="pa-5 text-center">
      <h5
        class="text-h5 mb-2 font-weight-bold"
        data-test="welcome-fourth-thanks"
      >
        Thank you for choosing ShellHub!
      </h5>
      <p class="text-body-1 mb-0">
        We're excited to have you on board. Happy connecting!
      </p>
    </div>
  </div>
</template>
