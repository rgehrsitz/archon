<template>
  <div class="h-100 d-flex align-center justify-space-between">
    <!-- 👉 Footer: left content -->
    <span class="d-flex align-center">
      &copy;
      {{ new Date().getFullYear() }}
      Made With
      <VIcon icon="bx-heart" color="error" size="1.25rem" class="mx-1" />
      By <a href="https://digwave.com" target="_blank" rel="noopener noreferrer" class="text-primary ms-1">Digital
        Wave</a>
    </span>
    <!-- 👉 Footer: right content -->
    <span class="d-md-flex gap-x-4 text-primary d-none">
      <a href="https://mit-license.org/" target="noopener noreferrer">License</a>
      <a href="https://xyzjdhfgsd.com" target="noopener noreferrer">Documentation</a>
    </span>
  </div>
</template>
